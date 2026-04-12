package tui

import (
	"fmt"
	"strings"

	"copytool/internal/theme"
	// "github.com/charmbracelet/lipgloss"
)

type filtersDrawerMetrics struct {
	PanelW      int
	PanelH      int
	InputW      int
	IncludeLine int
	ExcludeLine int
	DirsLine    int
	ClipLine    int
}

func (m Model) filtersDrawerSizing(th theme.Theme) filtersDrawerMetrics {
	panelW := min(92, max(58, m.Width-16))
	// Keep height large enough to avoid content overflow shifting mouse hit rows.
	const minPanelH = 32
	panelH := min(34, max(minPanelH, m.Height-6))

	// same as before; no button column
	inputW := max(24, panelW-10)

	return filtersDrawerMetrics{
		PanelW:      panelW,
		PanelH:      panelH,
		InputW:      inputW,
		IncludeLine: 5,
		ExcludeLine: 9,
		DirsLine:    13,
		ClipLine:    17,
	}
}

func (m Model) renderFiltersDrawer(th theme.Theme) string {
	metrics := m.filtersDrawerSizing(th)

	includeInput := m.FilterIncludeInput
	excludeInput := m.FilterExcludeInput
	dirsInput := m.FilterDirsInput

	includeInput.Width = metrics.InputW
	excludeInput.Width = metrics.InputW
	dirsInput.Width = metrics.InputW

	clipboardState := "[ ] off"
	if m.ClipboardEnabled {
		clipboardState = "[x] on"
	}

	lines := []string{
		th.HeaderTitle.Render("Filters"),
		"",
		th.Muted.Render("Refine the visible workspace, then apply and rescan the tree."),
		"",
		th.Accent.Render("Include extensions"),
		includeInput.View(),
		th.Muted.Render("  comma separated · empty means allow every extension"),
		"",
		th.Accent.Render("Exclude extensions"),
		excludeInput.View(),
		th.Muted.Render("  comma separated · matched files disappear from the tree"),
		"",
		th.Accent.Render("Exclude directories"),
		dirsInput.View(),
		th.Muted.Render("  comma separated · .git and tools remain hidden always"),
		"",
		th.Accent.Render("Clipboard"),
		th.Text.Render(fmt.Sprintf("  %s  export result to clipboard", clipboardState)),
		"",
		th.Accent.Render("Controls"),
		th.Muted.Render("  Click field        focus"),
		th.Muted.Render("  Click clipboard    toggle"),
		th.Muted.Render("  Tab / Shift+Tab    move between fields"),
		th.Muted.Render("  Enter              apply filters"),
		th.Muted.Render("  Ctrl+R             reset fields to defaults"),
		th.Muted.Render("  Ctrl+B             toggle clipboard"),
		th.Muted.Render("  Esc                close drawer"),
		th.Muted.Render("  plain typing always goes into the focused input"),
	}

	panel := th.PanelFocused.Copy().
		Width(metrics.PanelW).
		Height(metrics.PanelH).
		Render(strings.Join(lines, "\n"))

	return placeOverlay(m, max(40, m.Width-2), max(12, m.Height-2), panel)
}
