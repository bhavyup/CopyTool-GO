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

	controls := footerLabel(th, "COMPACT") + "  " + strings.Join([]string{
		footerCmd(th, "t", "oggle"),
		footerCmd(th, "s", "ave"),
		footerCmd(th, "c", "lipboard"),
		footerCmd(th, "f", "ilters"),
		footerCmd(th, "w", "orkspace"),
		footerCmd(th, "q", "uit"),
		footerCmd(th, "?", " help"),
	}, "   ")

	nav := footerLabel(th, "NAV") + "  " + strings.Join([]string{
		footerCmd(th, "tab", " pane"),
		footerCmd(th, "↑↓", " move"),
		footerCmd(th, "←→", " tree"),
		footerCmd(th, "PgUp/PgDn", " page"),
		footerCmd(th, "Home/End", " bounds"),
		footerCmd(th, "e", " expand"),
	}, "   ")

	status := footerLabel(th, "STATUS") + "  " + th.Status.Render("compact  ·  "+m.StatusMessage)

	return th.Footer.Render(
		rule + "\n" +
			footerFit(maxWidth, controls) + "\n" +
			footerFit(maxWidth, nav) + "\n" +
			footerFit(maxWidth, status),
	)
}
