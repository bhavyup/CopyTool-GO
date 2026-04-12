package tui

import (
	"fmt"
	"strings"

	"copytool/internal/core"

	"copytool/internal/theme"
	// "github.com/charmbracelet/lipgloss"
)

func (m Model) renderPalette(th theme.Theme) string {
	width := min(110, max(64, m.Width-10))
	height := min(32, max(16, m.Height-6))

	contentW := max(24, width-8)
	listH := max(6, height-10)

	header := th.HeaderTitle.Render("Search") + "  " + th.BadgeMuted.Render("fuzzy jump")
	sub := th.Muted.Render("Type to search paths. Enter jumps. Ctrl+T toggles selection. Esc closes.")

	// Input rendering
	in := m.PaletteInput
	in.Width = max(20, contentW-4)
	in.Prompt = ""
	inputLine := th.Key.Render("/") + " " + in.View()

	lines := []string{
		header,
		"",
		sub,
		"",
		inputLine,
		"",
	}

	if len(m.PaletteResults) == 0 {
		lines = append(lines, th.Muted.Render("No matches."))
	} else {
		start, end := windowRange(len(m.PaletteResults), m.PaletteIndex, listH)
		for i := start; i < end; i++ {
			item := m.PaletteResults[i]
			row := truncateRunes(item.Label, contentW)
			if i == m.PaletteIndex {
				lines = append(lines, th.RowFocused.Render(row))
			} else if item.IsDir {
				lines = append(lines, th.Accent.Render(row))
			} else {
				lines = append(lines, th.Text.Render(row))
			}
		}
	}

	lines = append(lines,
		"",
		th.Accent.Render("Controls"),
		th.Muted.Render("  ↑/↓      navigate"),
		th.Muted.Render("  Enter    jump to selection"),
		th.Muted.Render("  Ctrl+T   toggle selection"),
		th.Muted.Render("  Esc      close"),
	)

	panel := th.PanelFocused.Copy().
		Width(width).
		Height(height).
		Render(strings.Join(lines, "\n"))

	return placeOverlay(m, max(40, m.Width-2), max(12, m.Height-2), panel)
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
