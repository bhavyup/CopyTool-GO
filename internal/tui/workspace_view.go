package tui

import (
	"strings"

	"copytool/internal/theme"

	"github.com/charmbracelet/lipgloss"
)

type workspaceDrawerMetrics struct {
	PanelW     int
	PanelH     int
	InputW     int
	ButtonW    int
	RootLine   int // 0-based in content
	OutputLine int
	ClipLine   int
}

func (m Model) workspaceDrawerSizing(th theme.Theme) workspaceDrawerMetrics {
	panelW := min(98, max(60, m.Width-14))
	// Keep height large enough to avoid content overflow shifting mouse hit rows.
	const minPanelH = 29
	panelH := min(32, max(minPanelH, m.Height-6))

	btn := m.workspaceBrowsePill(th)
	btnW := lipgloss.Width(btn)

	// Keep this in sync with renderWorkspaceDrawer
	// Right side: " " + btn
	inputW := max(24, panelW-12-btnW-1)

	return workspaceDrawerMetrics{
		PanelW:     panelW,
		PanelH:     panelH,
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

	rootInput := m.WorkspaceRootInput
	outputInput := m.WorkspaceOutputInput
	rootInput.Width = metrics.InputW
	outputInput.Width = metrics.InputW

	btn := m.workspaceBrowsePill(th)

	rootLine := lipgloss.JoinHorizontal(lipgloss.Left, rootInput.View(), " ", btn)
	outputLine := lipgloss.JoinHorizontal(lipgloss.Left, outputInput.View(), " ", btn)

	clipboardState := "[ ] off"
	if m.ClipboardEnabled {
		clipboardState = "[x] on"
	}

	resolvedPreview, previewErr := m.workspaceResolvedOutputPreview()

	lines := []string{
		th.HeaderTitle.Render("Workspace"),
		"",
		th.Muted.Render("Change the active root and output destination without relaunching the command."),
		"",
		th.Accent.Render("Root path"),
		rootLine,
		th.Muted.Render("  changing root triggers an async rescan"),
		"",
		th.Accent.Render("Output target"),
		outputLine,
		th.Muted.Render("  empty = <root>/tools/output.txt · '.' = <cwd>/tools/output.txt"),
		"",
		th.Accent.Render("Resolved output preview"),
	}

	if previewErr != nil {
		lines = append(lines, th.Warning.Render("  "+previewErr.Error()))
	} else {
		lines = append(lines, th.Text.Render("  "+truncateMiddleRunes(resolvedPreview, metrics.InputW+metrics.ButtonW)))
	}

	lines = append(lines,
		"",
		th.Accent.Render("Clipboard"),
		th.Text.Render("  "+clipboardState+"  export result to clipboard"),
		"",
		th.Accent.Render("Controls"),
		th.Muted.Render("  Click field         focus"),
		th.Muted.Render("  Click ... / Ctrl+P  open picker for focused field"),
		th.Muted.Render("  Enter               apply workspace changes"),
		th.Muted.Render("  Ctrl+R              reset fields to current workspace"),
		th.Muted.Render("  Ctrl+B              toggle clipboard"),
		th.Muted.Render("  Esc                 close drawer"),
	)

	panel := th.PanelFocused.Copy().
		Width(metrics.PanelW).
		Height(metrics.PanelH).
		Render(strings.Join(lines, "\n"))

	// Keep Place sizing consistent with hit testing
	return placeOverlay(m, max(40, m.Width-2), max(12, m.Height-2), panel)
}
