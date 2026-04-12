package tui

import (
	"fmt"
	"strings"

	"copytool/internal/theme"

	"github.com/charmbracelet/lipgloss"
)

func (m Model) renderLoadingBody(th theme.Theme, height int) string {
	availableWidth := max(40, m.Width-4)
	gap := 1

	leftWidth := availableWidth * 56 / 100
	rightWidth := availableWidth - leftWidth - gap

	navigator := renderPanel(
		th.PanelFocused,
		th,
		"Navigator  • loading",
		strings.Join(m.loadingNavigatorLines(th, panelContentWidth(leftWidth), panelContentHeight(height)), "\n"),
		leftWidth,
		height,
	)

	inspect := renderPanel(
		th.Panel,
		th,
		"Inspect  • loading",
		strings.Join(m.loadingInspectLines(th, panelContentWidth(rightWidth), panelContentHeight(height)), "\n"),
		rightWidth,
		height,
	)

	return lipgloss.JoinHorizontal(
		lipgloss.Top,
		navigator,
		" ",
		inspect,
	)
}

func (m Model) renderScanErrorBody(th theme.Theme, height int) string {
	width := min(96, max(56, m.Width-12))

	lines := []string{
		th.HeaderTitle.Render("Project scan failed"),
		"",
		th.Warning.Render(truncateRunes(m.LoadError, width-8)),
		"",
		th.Accent.Render("Actions"),
		fmt.Sprintf("  %s  retry project scan", th.Key.Render("r")),
		fmt.Sprintf("  %s  quit", th.Key.Render("q")),
	}

	panel := th.PanelFocused.Copy().
		Width(width).
		Render(strings.Join(lines, "\n"))

	return placeOverlay(m, max(40, m.Width-2), max(8, height), panel)
}

func (m Model) loadingNavigatorLines(th theme.Theme, width, height int) []string {
	lines := []string{
		th.Muted.Render(m.LoadingLabel),
		"",
	}

	rows := max(6, min(height-2, 14))
	for i := 0; i < rows; i++ {
		depth := i % 4
		indent := strings.Repeat("  ", depth)

		prefix := "▸ "
		if i%5 == 4 {
			prefix = "· "
		}

		barWidth := max(8, width-lipgloss.Width(indent)-lipgloss.Width(prefix)-2)
		bar := renderSkeletonBar(th, barWidth, m.SkeletonPhase, i*3)

		lines = append(lines, indent+th.Muted.Render(prefix)+bar)
	}

	return clipLines(lines, height)
}

func (m Model) loadingInspectLines(th theme.Theme, width, height int) []string {
	lines := []string{
		renderSkeletonBar(th, max(14, width/2), m.SkeletonPhase, 0),
		renderSkeletonBar(th, max(20, width-6), m.SkeletonPhase, 3),
		"",
		th.Accent.Render("Selection"),
		"  " + renderSkeletonBar(th, max(12, width-8), m.SkeletonPhase, 5),
		"  " + renderSkeletonBar(th, max(10, width-12), m.SkeletonPhase, 7),
		"",
		th.Accent.Render("Metadata"),
		"  " + renderSkeletonBar(th, max(18, width-10), m.SkeletonPhase, 9),
		"  " + renderSkeletonBar(th, max(24, width-8), m.SkeletonPhase, 11),
		"  " + renderSkeletonBar(th, max(12, width-14), m.SkeletonPhase, 13),
		"",
		th.Accent.Render("Preview"),
	}

	previewRows := max(6, min(height-len(lines)-2, 12))
	for i := 0; i < previewRows; i++ {
		lineNo := fmt.Sprintf("%2d │", i+1)
		barWidth := max(8, width-lipgloss.Width(lineNo)-2)
		lines = append(lines, th.Meta.Render(lineNo)+" "+renderSkeletonBar(th, barWidth, m.SkeletonPhase, 17+i*2))
	}

	return clipLines(lines, height)
}

func renderSkeletonBar(th theme.Theme, width, phase, offset int) string {
	if width <= 0 {
		return ""
	}

	highlight := (phase*4 + offset) % max(1, width)

	var b strings.Builder
	for i := 0; i < width; i++ {
		switch {
		case i >= highlight && i < highlight+2:
			b.WriteString(th.Accent.Render("━"))
		case i >= highlight-2 && i < highlight+5:
			b.WriteString(th.Text.Render("━"))
		default:
			b.WriteString(th.Muted.Render("━"))
		}
	}

	return b.String()
}

func clipLines(lines []string, maxLines int) []string {
	if maxLines <= 0 {
		return []string{}
	}

	if len(lines) <= maxLines {
		return lines
	}

	clipped := append([]string{}, lines[:maxLines]...)
	clipped[maxLines-1] = "…"
	return clipped
}
