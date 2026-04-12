package tui

import (
	"strings"

	"copytool/internal/theme"
	// "github.com/charmbracelet/lipgloss"
)

func (m Model) renderPicker(th theme.Theme) string {
	width := min(100, max(60, m.Width-12))
	height := min(28, max(14, m.Height-8))
	contentWidth := max(24, width-8)
	contentHeight := max(8, height-8)

	title := "Picker"
	if m.PickerMode == "root" {
		title = "Picker  • root"
	} else if m.PickerMode == "output" {
		title = "Picker  • output"
	}

	lines := []string{
		th.HeaderTitle.Render(title),
		"",
		th.Text.Render("path"),
		th.Muted.Render(truncateMiddleRunes(m.PickerPath, contentWidth)),
		"",
	}

	start, end := windowRange(len(m.PickerEntries), m.PickerIndex, contentHeight)

	for i := start; i < end; i++ {
		entry := m.PickerEntries[i]
		lines = append(lines, m.renderPickerRow(th, entry, i == m.PickerIndex, contentWidth))
	}

	lines = append(lines,
		"",
		th.Accent.Render("Controls"),
		th.Muted.Render("  ↑ / ↓        move"),
		th.Muted.Render("  Enter        open directory / choose file"),
		th.Muted.Render("  Space        choose highlighted path"),
		th.Muted.Render("  Backspace/←  parent directory"),
		th.Muted.Render("  Esc          close picker"),
	)

	panel := th.PanelFocused.Copy().
		Width(width).
		Height(height).
		Render(strings.Join(lines, "\n"))

	return placeOverlay(m, max(40, m.Width-2), max(12, m.Height-2), panel)
}

func (m Model) renderPickerRow(th theme.Theme, entry pickerEntry, focused bool, width int) string {
	prefix := "· "
	switch entry.Kind {
	case "current":
		prefix = "◎ "
	case "parent":
		prefix = "↩ "
	case "dir":
		prefix = "▸ "
	case "file":
		prefix = "• "
	}

	text := truncateRunes(prefix+entry.Label, width)

	if focused {
		return th.RowFocused.Render(text)
	}

	if entry.IsDir {
		return th.Accent.Render(text)
	}

	return th.Text.Render(text)
}
