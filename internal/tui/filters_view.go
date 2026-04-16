package tui

import (
	"strings"

	"copytool/internal/theme"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
)

type filtersDrawerMetrics struct {
	PanelW      int
	PanelH      int
	ContentW    int
	InputW      int
	IncludeLine int
	ExcludeLine int
	DirsLine    int
	ClipLine    int
}

type filtersDrawerLayout struct {
	Lines            []string
	IncludeLineIndex int
	ExcludeLineIndex int
	DirsLineIndex    int
	ClipLineIndex    int
}

func (m Model) filtersDrawerSizing(th theme.Theme) filtersDrawerMetrics {
	panelW := min(110, max(68, m.Width-12))
	// Match panel height to current rendered content to avoid dead space at bottom.
	const contentLines = 23
	panelH := min(contentLines, max(16, m.Height-6))

	contentW := max(30, panelW-th.PanelFocused.GetHorizontalFrameSize())
	inputW := max(24, contentW-1)

	return filtersDrawerMetrics{
		PanelW:      panelW,
		PanelH:      panelH,
		ContentW:    contentW,
		InputW:      inputW,
		IncludeLine: 5,
		ExcludeLine: 9,
		DirsLine:    13,
		ClipLine:    17,
	}
}

func (m Model) renderFiltersDrawer(th theme.Theme) string {
	metrics := m.filtersDrawerSizing(th)
	layout := m.filtersDrawerLayout(th, metrics)

	panel := th.PanelFocused.Copy().
		Width(metrics.PanelW).
		Height(metrics.PanelH).
		Render(strings.Join(layout.Lines, "\n"))

	return placeOverlay(m, max(40, m.Width-2), max(12, m.Height-2), panel)
}

func (m Model) filtersDrawerLayout(th theme.Theme, metrics filtersDrawerMetrics) filtersDrawerLayout {

	includeInput := m.FilterIncludeInput
	excludeInput := m.FilterExcludeInput
	dirsInput := m.FilterDirsInput

	includeInput.Width = metrics.InputW
	excludeInput.Width = metrics.InputW
	dirsInput.Width = metrics.InputW

	inputRowStyle := th.DrawerInput.Copy().
		Width(metrics.ContentW).
		Padding(0, 0)

	includeRow := inputRowStyle.Render(includeInput.View())
	excludeRow := inputRowStyle.Render(excludeInput.View())
	dirsRow := inputRowStyle.Render(dirsInput.View())

	focusLabel := m.filtersFocusLabel()

	profileChip := th.HeaderBadge.Render("base")
	if m.hasCustomFilters() {
		profileChip = th.HeaderBadge.Copy().Foreground(th.P.Accent).Bold(true).Render("custom")
	}

	clipboardGlyph := th.TreeSelNone.Render("○")
	clipboardChip := clipboardGlyph + th.HeaderBadge.Background(th.P.Surface).Render("clipboard")
	if m.ClipboardEnabled {
		clipboardGlyph = th.TreeSelFull.Render("●")
		clipboardChip = clipboardGlyph + th.HeaderBadge.Copy().Background(th.P.Surface).Foreground(th.P.Success).Bold(true).Render("clipboard")
	}

	rightChips := strings.Join([]string{
		profileChip,
		th.HeaderBadge.Render("focus: " + focusLabel),
		clipboardChip,
	}, th.TextWS.Render(" ▏"))

	titleRow := padBetween(th, th.TreeRowDir.Render("Filters"), rightChips, metrics.ContentW)
	rule := th.Muted.Background(th.P.Surface).Render(strings.Repeat("─", metrics.ContentW))

	controlsTop := []string{
		theme.KeyHintAccent(th, "Enter", " apply"),
		FilterFooterCmd(th, "Esc", " close"),
		FilterFooterCmd(th, "Tab", " next"),
		FilterFooterCmd(th, "Shift+Tab", " prev"),
	}

	controlsBottom := []string{
		FilterFooterCmd(th, "Ctrl+R", " reset"),
		FilterFooterCmd(th, "Ctrl+B", " clipboard"),
		FilterFooterCmd(th, "click", " focus/toggle"),
	}

	controlsPrimary, controlsSecondary := renderFilterControlRows(th, metrics.ContentW, controlsTop, controlsBottom)

	lines := []string{
		titleRow,
		rule,
		th.DrawerHint.Render("Edit filter sets and apply to rescan the workspace tree."),
		"",
		theme.RuledSection(th, "INCLUDE EXTENSIONS", metrics.ContentW),
	}

	includeLineIndex := len(lines)
	lines = append(lines,
		includeRow,
		RenderAtEdge(th, th.DrawerHint.Render("comma separated, example: go,md,txt"), metrics.ContentW),
		"",
		theme.RuledSection(th, "EXCLUDE EXTENSIONS", metrics.ContentW),
	)

	excludeLineIndex := len(lines)
	lines = append(lines,
		excludeRow,
		RenderAtEdge(th, th.DrawerHint.Render("comma separated, example: log,tmp"), metrics.ContentW),
		"",
		theme.RuledSection(th, "EXCLUDE DIRECTORIES", metrics.ContentW),
	)

	dirsLineIndex := len(lines)
	lines = append(lines,
		dirsRow,
		RenderAtEdge(th, th.DrawerHint.Render("comma separated, .git & tools remain hidden always"), metrics.ContentW),
		"",
		theme.RuledSection(th, "OUTPUT", metrics.ContentW),
	)

	clipLineIndex := len(lines)
	lines = append(lines,
		th.Text.Background(th.P.Surface).Render("  ")+clipboardGlyph+th.Text.Background(th.P.Surface).Render("  export result to clipboard"),
		RenderAtEdge(th, th.DrawerHint.Render("toggled with Ctrl+B/mouse click"), metrics.ContentW),
		"",
		theme.RuledSection(th, "CONTROLS", metrics.ContentW),
		controlsPrimary,
		controlsSecondary,
	)

	return filtersDrawerLayout{
		Lines:            lines,
		IncludeLineIndex: includeLineIndex,
		ExcludeLineIndex: excludeLineIndex,
		DirsLineIndex:    dirsLineIndex,
		ClipLineIndex:    clipLineIndex,
	}
}

func (m Model) filtersFocusLabel() string {
	switch m.FilterFocusIndex {
	case 1:
		return "exclude ext"
	case 2:
		return "exclude dirs"
	default:
		return "include ext"
	}
}

func RenderAtEdge(th theme.Theme, label string, width int) string {
	padding := max(0, width-lipgloss.Width(label))
	return th.Text.Render(strings.Repeat(" ", padding) + label)
}

func FilterFooterCmd(th theme.Theme, key, desc string) string {
	keyPart := th.TreeRowDir.Bold(true).Background(lipgloss.Color("#000000")).Render(key)
	textPart := th.Status.Render(desc)
	return keyPart + textPart
}

func renderFilterControlRows(th theme.Theme, width int, top, bottom []string) (string, string) {
	cols := max(len(top), len(bottom))
	if cols == 0 {
		return "", ""
	}

	colWidths := make([]int, cols)
	base := max(1, width/cols)
	rem := max(0, width-base*cols)

	for i := 0; i < cols; i++ {
		colWidths[i] = base
		if i < rem {
			colWidths[i]++
		}
	}

	return renderFilterControlRow(th, width, top, colWidths), renderFilterControlRow(th, width, bottom, colWidths)
}

func renderFilterControlRow(th theme.Theme, width int, items []string, colWidths []int) string {
	var b strings.Builder
	for i, cellWidth := range colWidths {
		item := ""
		if i < len(items) {
			item = items[i]
		}

		cell := ansi.Truncate(item, max(1, cellWidth), "")
		if i < len(colWidths)-1 {
			cellW := lipgloss.Width(cell)
			if cellW < cellWidth {
				cell += strings.Repeat(th.TextWS.Render(" "), cellWidth-cellW)
			}
		}

		b.WriteString(cell)
	}

	return ansi.Truncate(b.String(), width, "")
}
