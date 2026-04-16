package tui

import (
	"path/filepath"
	"strings"

	"copytool/internal/theme"

	"github.com/charmbracelet/lipgloss"
)

type workspaceDrawerMetrics struct {
	PanelW     int
	PanelH     int
	ContentW   int
	InputW     int
	ButtonW    int
	RootLine   int // 0-based in content
	OutputLine int
	ClipLine   int
}

type workspaceDrawerLayout struct {
	Lines             []string
	RootLineIndex     int
	OutputLineIndex   int
	ClipLineIndex     int
	RootButtonStart   int
	RootButtonEnd     int
	OutputButtonStart int
	OutputButtonEnd   int
}

func (m Model) workspaceDrawerSizing(th theme.Theme) workspaceDrawerMetrics {
	panelW := min(110, max(68, m.Width-12))
	const contentLines = 22
	panelH := min(contentLines, max(16, m.Height-6))

	btn := m.workspaceBrowsePill(th)
	btnW := lipgloss.Width(btn)
	contentW := max(30, panelW-th.PanelFocused.GetHorizontalFrameSize())
	const inputReserve = 1

	// Keep this in sync with renderWorkspaceDrawer
	// Right side: " " + btn
	inputW := max(24, contentW-btnW-1-inputReserve)

	return workspaceDrawerMetrics{
		PanelW:     panelW,
		PanelH:     panelH,
		ContentW:   contentW,
		InputW:     inputW,
		ButtonW:    btnW,
		RootLine:   5,
		OutputLine: 9,
		ClipLine:   16,
	}
}

func (m Model) workspaceBrowsePill(th theme.Theme) string {
	// A nicer "..." pill (still compact), using existing theme surfaces.
	return th.Key.Copy().Padding(0, 2).Render("...")
}

func (m Model) renderWorkspaceDrawer(th theme.Theme) string {
	metrics := m.workspaceDrawerSizing(th)
	layout := m.workspaceDrawerLayout(th, metrics)

	panel := th.PanelFocused.Copy().
		Width(metrics.PanelW).
		Height(metrics.PanelH).
		Render(strings.Join(layout.Lines, "\n"))

	// Keep Place sizing consistent with hit testing
	return placeOverlay(m, max(40, m.Width-2), max(12, m.Height-2), panel)
}

func (m Model) workspaceDrawerLayout(th theme.Theme, metrics workspaceDrawerMetrics) workspaceDrawerLayout {
	rootInput := m.WorkspaceRootInput
	outputInput := m.WorkspaceOutputInput
	rootInput.Width = metrics.InputW
	outputInput.Width = metrics.InputW

	btn := m.workspaceBrowsePill(th)
	inputRowStyle := th.DrawerInput.Copy().Width(metrics.ContentW).Padding(0, 0)
	btnW := lipgloss.Width(btn)
	rootInputView := rootInput.View()
	outputInputView := outputInput.View()

	rootLine := lipgloss.JoinHorizontal(lipgloss.Left, rootInputView, " ", btn)
	outputLine := lipgloss.JoinHorizontal(lipgloss.Left, outputInputView, "  ", btn) //deliberately extra space to align with root line (since output input is shorter due to btn)
	rootRow := inputRowStyle.Render(rootLine)
	outputRow := inputRowStyle.Render(outputLine)

	clipboardGlyph := th.TreeSelNone.Render("○")
	clipboardChip := clipboardGlyph + th.HeaderBadge.Background(th.P.Surface).Render("clipboard")
	if m.ClipboardEnabled {
		clipboardGlyph = th.TreeSelFull.Render("●")
		clipboardChip = clipboardGlyph + th.HeaderBadge.Copy().Background(th.P.Surface).Foreground(th.P.Success).Bold(true).Render("clipboard")
	}

	resolvedPreview, previewErr := m.workspaceResolvedOutputPreview()

	focusChip := th.HeaderBadge.Render("focus: " + m.workspaceFocusLabel())
	rootChip := th.HeaderBadge.Render("root: " + filepath.Base(m.RootPath))
	titleRight := strings.Join([]string{rootChip, focusChip, clipboardChip}, th.TextWS.Render(" ▏"))

	titleRow := padBetween(th, th.TreeRowDir.Render("Workspace"), titleRight, metrics.ContentW)
	rule := th.Muted.Background(th.P.Surface).Render(strings.Repeat("─", metrics.ContentW))

	previewText := resolvedPreview
	previewStyle := th.Text
	if previewErr != nil {
		previewText = "invalid output: " + previewErr.Error()
		previewStyle = th.Warning
	}

	previewRow := inputRowStyle.Render(previewStyle.Render(" " + truncateMiddleRunes(previewText, max(1, metrics.ContentW-1))))

	controlsTop := []string{
		theme.KeyHintAccent(th, "Enter", " apply"),
		FilterFooterCmd(th, "Esc", " close"),
		FilterFooterCmd(th, "Tab", " next"),
		FilterFooterCmd(th, "Shift+Tab", " prev"),
	}

	controlsBottom := []string{
		FilterFooterCmd(th, "Ctrl+R", " reset"),
		FilterFooterCmd(th, "Ctrl+P", " browse"),
		FilterFooterCmd(th, "Ctrl+B", " clipboard"),
		FilterFooterCmd(th, "click", " focus/toggle"),
	}

	controlsPrimary, controlsSecondary := renderFilterControlRows(th, metrics.ContentW, controlsTop, controlsBottom)

	lines := []string{
		titleRow,
		rule,
		th.DrawerHint.Render("Change root and output targets without leaving the current session."),
		"",
		theme.RuledSection(th, "ROOT PATH", metrics.ContentW),
	}

	rootLineIndex := len(lines)
	lines = append(lines,
		rootRow,
		RenderAtEdge(th, th.DrawerHint.Render("changing root triggers async rescan"), metrics.ContentW),
		"",
		theme.RuledSection(th, "OUTPUT TARGET", metrics.ContentW),
	)

	outputLineIndex := len(lines)
	lines = append(lines,
		outputRow,
		RenderAtEdge(th, th.DrawerHint.Render("empty = <root>/tools/output.txt  ·  . = <cwd>/tools/output.txt"), metrics.ContentW),
		"",
		theme.RuledSection(th, "RESOLVED OUTPUT", metrics.ContentW),
		previewRow,
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

	rootButtonStart := lipgloss.Width(rootInputView) + 1
	rootButtonEnd := rootButtonStart + btnW
	outputButtonStart := lipgloss.Width(outputInputView) + 1
	outputButtonEnd := outputButtonStart + btnW

	return workspaceDrawerLayout{
		Lines:             lines,
		RootLineIndex:     rootLineIndex,
		OutputLineIndex:   outputLineIndex,
		ClipLineIndex:     clipLineIndex,
		RootButtonStart:   rootButtonStart,
		RootButtonEnd:     rootButtonEnd,
		OutputButtonStart: outputButtonStart,
		OutputButtonEnd:   outputButtonEnd,
	}
}

func (m Model) workspaceFocusLabel() string {
	if m.WorkspaceFocusIndex == 1 {
		return "output target"
	}
	return "root path"
}
