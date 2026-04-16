package tui

import (
	"strings"

	"copytool/internal/theme"
)

func (m Model) isCompactLayout() bool {
	return m.Width < 72 || m.Height < 18
}

func (m Model) renderCompactBody(th theme.Theme, height int) string {
	width := max(20, m.Width-4)

	if m.ActivePane == paneInspect {
		return renderPanel(
			th.PanelFocused,
			th,
			"Inspect  • compact",
			strings.Join(
				m.renderInspectLines(th, panelContentWidth(width), panelContentHeight(height)),
				"\n",
			),
			width,
			height,
		)
	}

	return renderPanel(
		th.PanelFocused,
		th,
		"Navigator  • compact",
		strings.Join(
			m.renderTreeLines(th, panelContentWidth(width), panelContentHeight(height)),
			"\n",
		),
		width,
		height,
	)
}

func (m Model) renderCompactFooter(th theme.Theme) string {
	maxWidth := max(20, m.Width-th.App.GetHorizontalFrameSize())
	rule := footerRule(th, maxWidth)

	actions := []string{
		footerCmdPrimary(th, "s", "save"),
		footerCmd(th, "q", "quit"),
		footerCmd(th, "t", "toggle"),
		footerCmd(th, "c", "clipboard"),
		footerCmd(th, "f", "filters"),
		footerCmd(th, "w", "workspace"),
		footerCmd(th, "?", "help"),
	}

	nav := []string{
		footerCmd(th, "tab", "pane"),
		footerCmd(th, "↑↓", "move"),
		footerCmd(th, "←→", "tree"),
		footerCmd(th, "PgUp/PgDn", "page"),
		footerCmd(th, "Home/End", "bounds"),
		footerCmd(th, "e", "expand"),
	}
	labelColWidth := footerLabelColumnWidth(th, "actions", "nav")
	colWidths := footerColumnWidths(actions, nav)

	rows := []string{
		rule,
		m.footerStatusLine(th, maxWidth, "compact"),
	}
	rows = append(rows, footerAlignedHintsLine(th, maxWidth, "actions", actions, colWidths, labelColWidth))
	rows = append(rows, footerAlignedHintsLine(th, maxWidth, "nav", nav, colWidths, labelColWidth))

	return th.Footer.Render(strings.Join(rows, "\n"))
}
