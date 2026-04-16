package tui

import (
	"fmt"
	"strings"

	"copytool/internal/core"

	"copytool/internal/theme"

	"github.com/charmbracelet/lipgloss"
)

func (m Model) renderPalette(th theme.Theme) string {
	panelStyle := th.PanelFocused.Copy()
	panelW := min(122, max(76, m.Width-10))
	panelH := min(30, max(18, m.Height-6))
	contentW := max(28, panelW-panelStyle.GetHorizontalFrameSize())

	query := strings.TrimSpace(m.PaletteInput.Value())
	queryChip := th.HeaderBadge.Render("query: empty")
	if query != "" {
		queryChip = th.HeaderBadge.Copy().Foreground(th.P.Accent).Bold(true).Render("query: active")
	}

	focusChip := th.HeaderBadge.Render("focus: none")
	if len(m.PaletteResults) > 0 {
		focusChip = th.HeaderBadge.Render(fmt.Sprintf("focus: %d/%d", m.PaletteIndex+1, len(m.PaletteResults)))
	}

	titleRight := strings.Join([]string{
		th.HeaderBadge.Render("index: fuzzy"),
		th.HeaderBadge.Render(fmt.Sprintf("results: %d/%d", len(m.PaletteResults), len(m.PaletteAll))),
		focusChip,
		queryChip,
	}, th.TextWS.Render(" ▏"))

	titleRow := padBetween(th, th.TreeRowDir.Render("Search"), titleRight, contentW)
	rule := th.Muted.Background(th.P.Surface).Render(strings.Repeat("─", contentW))

	in := m.PaletteInput
	in.Prompt = ""
	prefix := th.KeyChipAccent.Render("/")
	in.Width = max(16, contentW-lipgloss.Width(prefix)-1)
	inputRow := th.DrawerInput.Copy().
		Width(contentW).
		Padding(0, 0).
		Render(prefix + in.View())

	fixedBeforeResults := 8
	fixedAfterResults := 5
	listH := max(4, panelH-fixedBeforeResults-fixedAfterResults)

	lines := []string{
		titleRow,
		rule,
		th.DrawerHint.Render("Fuzzy-search any path and jump instantly to the focused match."),
		"",
		theme.RuledSection(th, "QUERY", contentW),
		inputRow,
		"",
		theme.RuledSection(th, "RESULTS", contentW),
	}

	if len(m.PaletteResults) == 0 {
		lines = append(lines, th.InspectMuted.Render("  no matches for current query"))
	} else {
		start, end := windowRange(len(m.PaletteResults), m.PaletteIndex, listH)
		for i := start; i < end; i++ {
			item := m.PaletteResults[i]

			selected := false
			partial := false
			if m.Tree != nil && m.Tree.Root != nil {
				node := core.FindNodeByRelPath(m.Tree.Root, item.RelPath)
				if node != nil {
					selected = node.Selected
					partial = node.Partial
				}
			}

			lines = append(lines, renderPaletteResultRow(th, item, i == m.PaletteIndex, selected, partial, contentW))
		}
	}

	controlsTop := []string{
		theme.KeyHintAccent(th, "Enter", " jump"),
		FilterFooterCmd(th, "Ctrl+T", " toggle"),
		FilterFooterCmd(th, "Esc", " close"),
		FilterFooterCmd(th, "↑/↓", " navigate"),
	}
	controlsBottom := []string{
		FilterFooterCmd(th, "j/k", " navigate"),
		FilterFooterCmd(th, "type", " filter query"),
	}
	controlsRow1, controlsRow2 := renderFilterControlRows(th, contentW, controlsTop, controlsBottom)

	lines = append(lines,
		"",
		theme.RuledSection(th, "CONTROLS", contentW),
		controlsRow1,
		controlsRow2,
	)

	panel := panelStyle.
		Width(panelW).
		Height(panelH).
		Render(strings.Join(lines, "\n"))

	return placeOverlay(m, max(40, m.Width-2), max(12, m.Height-2), panel)
}

func renderPaletteResultRow(th theme.Theme, item paletteItem, focused bool, selected bool, partial bool, width int) string {
	focusMark := th.TreeIndent.Render(" ")
	if focused {
		focusMark = th.KeyChipAccent.Render("▶")
	}

	iconStyle := th.TreeIndent
	nameStyle := th.TreeRowFile
	wsStyle := th.TextWS
	kindLabel := "file"
	kindStyle := th.HeaderBadge.Copy().Background(th.P.Surface)

	if item.IsDir {
		iconStyle = th.TreeDirClosed
		nameStyle = th.TreeRowDir
		kindLabel = "dir"
		kindStyle = kindStyle.Copy().Foreground(th.P.DirColor)
	}

	if focused {
		iconStyle = iconStyle.Copy().Background(th.P.Raised)
		nameStyle = nameStyle.Copy().Background(th.P.Raised).Bold(true)
		wsStyle = wsStyle.Copy().Background(th.P.Raised)
		kindStyle = kindStyle.Copy().Background(th.P.Raised).Foreground(th.P.TextPrimary)
	}

	kindChip := kindStyle.Render(kindLabel)

	icon := iconStyle.Render("·")
	if item.IsDir {
		icon = iconStyle.Render("▸")
	}

	selGlyph := theme.SelectionGlyph(th, selected, partial, focused)
	prefix := focusMark + wsStyle.Render(" ") + icon + wsStyle.Render(" ") + selGlyph + wsStyle.Render(" ")

	nameW := max(1, width-lipgloss.Width(prefix)-lipgloss.Width(kindChip)-1)
	name := nameStyle.Render(truncateRunes(item.Label, nameW))
	row := padBetween(th, prefix+name, kindChip, width)

	if focused {
		return th.TreeRowFocused.Copy().Width(width).Render(row)
	}

	return th.TreeRowNormal.Copy().Width(width).Render(row)
}

func (m *Model) paletteJumpToHighlighted() {
	if m.Tree == nil || m.Tree.Root == nil || len(m.PaletteResults) == 0 {
		return
	}

	item := m.PaletteResults[m.PaletteIndex]
	node := core.FindNodeByRelPath(m.Tree.Root, item.RelPath)
	if node == nil {
		return
	}

	core.ExpandAncestors(node)
	m.Tree.RebuildVisible()
	m.Tree.FocusNode(node)
	m.refreshPreview()
	m.StatusMessage = fmt.Sprintf("Jumped to: %s", item.RelPath)
}
