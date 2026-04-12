package tui

import (
	"fmt"
	"path/filepath"
	"strings"

	"copytool/internal/core"
	"copytool/internal/theme"

	"github.com/charmbracelet/lipgloss"
)

type inspectLineKind int

const (
	lineBlank inspectLineKind = iota
	lineTitle
	lineSubtitle
	lineSection
	lineText
	lineMuted
	lineWarning
	linePreview
	linePreviewNumbered
)

type inspectLine struct {
	Kind   inspectLineKind
	Text   string
	Number int
}

func (m Model) View() string {
	th := theme.New()

	if m.Width == 0 || m.Height == 0 {
		return th.App.Render(th.Muted.Render("initializing copytool..."))
	}

	header := m.renderHeader(th)
	footer := m.renderFooter(th)
	bodyHeight := max(1, m.Height-lipgloss.Height(header)-lipgloss.Height(footer))

	var body string
	switch {
	case m.LoadError != "":
		body = m.renderScanErrorBody(th, bodyHeight)
	case m.Loading:
		body = m.renderLoadingBody(th, bodyHeight)
	case m.isCompactLayout():
		body = m.renderCompactBody(th, bodyHeight)
	default:
		body = m.renderBody(th, bodyHeight)
	}

	ui := lipgloss.JoinVertical(
		lipgloss.Left,
		header,
		body,
		footer,
	)

	if m.ShowHelp {
		return m.renderHelpView(th)
	}

	if m.ShowQuitConfirm {
		return m.renderQuitConfirmView(th)
	}

	if m.ShowSaveSuccess {
		return m.renderSaveSuccessView(th)
	}

	if m.ShowPalette {
		return m.renderPalette(th)
	}

	if m.ShowPicker {
		return m.renderPicker(th)
	}

	if m.ShowWorkspace {
		return m.renderWorkspaceDrawer(th)
	}

	if m.ShowFilters {
		return m.renderFiltersDrawer(th)
	}

	return th.App.Render(ui)
}

func (m Model) renderHeader(th theme.Theme) string {
	project := filepath.Base(m.RootPath)
	appContentWidth := max(20, m.Width-th.App.GetHorizontalFrameSize())
	headerStyle := th.PanelFocused.Copy()
	headerContentWidth := max(1, appContentWidth-headerStyle.GetHorizontalFrameSize())
	infoWidth := max(20, headerContentWidth-22)

	var stats core.Stats
	var visibleCount int
	var focusedPath string
	selectedFiles := 0

	if m.Tree != nil && m.Tree.Root != nil {
		stats = core.ComputeStats(m.Tree.Root)
		visibleCount = len(m.Tree.VisibleNodes)
		selectedFiles = core.CountSelectedFiles(m.Tree.Root)

		if focused := m.Tree.FocusedNode(); focused != nil {
			focusedPath = focused.RelPath
		}
	} else if m.Loading {
		focusedPath = "scanning..."
	} else if m.LoadError != "" {
		focusedPath = "scan failed"
	}

	clipboardLabel := "clipboard off"
	if m.ClipboardEnabled {
		clipboardLabel = "clipboard on"
	}

	saveLabel := "saved"
	if m.Dirty {
		saveLabel = "draft"
	}

	items := []string{
		th.HeaderTitle.Render("COPYTOOL"),
		"  ",
		th.Badge.Render(project),
		" ",
		th.BadgeMuted.Render(fmt.Sprintf("%d selected", selectedFiles)),
		" ",
		th.BadgeMuted.Render(fmt.Sprintf("%d files", stats.Files)),
		" ",
		th.BadgeMuted.Render(fmt.Sprintf("%d folders", stats.Dirs)),
		" ",
		th.BadgeMuted.Render(fmt.Sprintf("%d visible", visibleCount)),
		" ",
		th.BadgeMuted.Render(m.filterBadgeLabel()),
		" ",
		th.BadgeMuted.Render(clipboardLabel),
		" ",
		th.BadgeMuted.Render(saveLabel),
	}

	if m.Loading {
		items = append(items, " ", th.BadgeMuted.Render("scanning"))
	} else if m.LoadError != "" {
		items = append(items, " ", th.BadgeMuted.Render("scan failed"))
	}

	titleRow := lipgloss.JoinHorizontal(lipgloss.Left, items...)

	details := []string{
		th.Text.Render("root") + "      " + th.Muted.Render(truncateMiddleRunes(m.RootPath, infoWidth)),
		th.Text.Render("output") + "    " + th.Muted.Render(truncateMiddleRunes(emptyFallback(m.ResolvedOutputPath, "none"), infoWidth)),
		th.Text.Render("focus") + "     " + th.Muted.Render(truncateMiddleRunes(emptyFallback(focusedPath, "none"), infoWidth)),
		th.Text.Render("seed") + "      " + th.Muted.Render(truncateRunes(summarizePaths(m.InitialSelections, 3), infoWidth)),
	}

	content := titleRow + "\n" + strings.Join(details, "\n")

	rendered := headerStyle.Copy().
		Width(headerContentWidth).
		Render(content)

	if used := lipgloss.Width(rendered); used < appContentWidth {
		headerContentWidth += appContentWidth - used
		rendered = headerStyle.Copy().
			Width(headerContentWidth).
			Render(content)
	}

	return rendered
}

func (m Model) renderBody(th theme.Theme, height int) string {
	availableWidth := max(40, m.Width-th.App.GetHorizontalFrameSize())
	gap := 1

	leftWidth := (availableWidth - gap) * 56 / 100
	rightWidth := availableWidth - leftWidth - gap

	navigator := m.renderNavigatorPanel(th, leftWidth, height)
	inspect := m.renderInspectPanel(th, rightWidth, height)

	return lipgloss.JoinHorizontal(
		lipgloss.Top,
		navigator,
		" ",
		inspect,
	)
}

func (m Model) renderNavigatorPanel(th theme.Theme, width, height int) string {
	contentWidth := panelContentWidth(width)
	contentHeight := panelContentHeight(height)

	bodyLines := m.renderTreeLines(th, contentWidth, contentHeight)
	body := strings.Join(bodyLines, "\n")

	style := th.Panel
	title := "Navigator"
	if m.ActivePane == paneTree {
		style = th.PanelFocused
		title = "Navigator • active"
	}

	return renderPanel(style, th, title, body, width, height)
}

func (m Model) renderInspectPanel(th theme.Theme, width, height int) string {
	contentWidth := panelContentWidth(width)
	contentHeight := panelContentHeight(height)

	bodyLines := m.renderInspectLines(th, contentWidth, contentHeight)
	body := strings.Join(bodyLines, "\n")

	style := th.Panel
	title := "Inspect"

	if m.ActivePane == paneInspect {
		style = th.PanelFocused
	}

	if m.ActivePane == paneInspect {
		title += "  • active"
	}

	return renderPanel(style, th, title, body, width, height)
}

func (m Model) renderTreeLines(th theme.Theme, width, height int) []string {
	if m.Tree == nil || len(m.Tree.VisibleNodes) == 0 {
		return []string{th.Muted.Render("No nodes available.")}
	}

	start, end := windowRange(len(m.Tree.VisibleNodes), m.Tree.FocusIndex, height)

	lines := make([]string, 0, end-start)
	for i := start; i < end; i++ {
		node := m.Tree.VisibleNodes[i]
		lines = append(lines, m.renderTreeRow(th, node, i == m.Tree.FocusIndex, width))
	}

	if len(lines) == 0 {
		return []string{th.Muted.Render("No visible rows.")}
	}

	return lines
}

func (m Model) renderTreeRow(th theme.Theme, node *core.Node, focused bool, width int) string {
	depth := node.Depth()
	indent := strings.Repeat("  ", depth)

	marker := "·"
	if node.IsDir {
		switch {
		case len(node.Children) == 0:
			marker = "▹"
		case node.Expanded:
			marker = "▾"
		default:
			marker = "▸"
		}
	}

	checkbox := checkboxForNode(node)

	label := node.Name
	if node.IsDir {
		label += "/"
	}

	meta := ""
	if node.IsDir {
		selectedInside := core.CountSelectedFiles(node)
		meta = fmt.Sprintf("%d child · %d sel", len(node.Children), selectedInside)
	} else if node.Ext != "" {
		meta = strings.TrimPrefix(node.Ext, ".")
	}

	rowWidth := width
	if focused {
		rowWidth = max(1, width-th.RowFocused.GetHorizontalFrameSize())
	}

	plain := renderTreeRowJustified(indent+marker+" "+checkbox+" ", label, meta, rowWidth)

	if focused {
		return th.RowFocused.Render(plain)
	}

	if node.IsDir {
		return th.Accent.Render(plain)
	}

	return th.Text.Render(plain)
}

func renderTreeRowJustified(prefix, label, meta string, width int) string {
	if width <= 0 {
		return ""
	}

	left := prefix + label
	if meta == "" {
		return truncateRunes(left, width)
	}

	rightWidth := lipgloss.Width(meta)
	if rightWidth >= width {
		return truncateRunes(meta, width)
	}

	minGap := 1
	maxLeftWidth := width - rightWidth - minGap
	if maxLeftWidth <= 0 {
		return truncateRunes(meta, width)
	}

	left = truncateRunes(left, maxLeftWidth)
	gap := width - lipgloss.Width(left) - rightWidth
	if gap < minGap {
		gap = minGap
	}

	return left + strings.Repeat(" ", gap) + meta
}

func (m Model) buildInspectLineItems() []inspectLine {
	if m.Tree == nil {
		return []inspectLine{
			{Kind: lineMuted, Text: "Tree not loaded."},
		}
	}

	node := m.Tree.FocusedNode()
	if node == nil {
		return []inspectLine{
			{Kind: lineMuted, Text: "Nothing focused."},
		}
	}

	stats := core.ComputeStats(node)
	selectedFilesInNode := core.CountSelectedFiles(node)
	items := []inspectLine{
		{Kind: lineTitle, Text: node.Name},
		{Kind: lineSubtitle, Text: node.RelPath},
		{Kind: lineBlank},
		{Kind: lineSection, Text: "Selection"},
		{Kind: lineText, Text: fmt.Sprintf("state          %s", core.SelectionState(node))},
		{Kind: lineText, Text: fmt.Sprintf("selected files %d", selectedFilesInNode)},
		{Kind: lineBlank},
		{Kind: lineSection, Text: "Metadata"},
		{Kind: lineText, Text: fmt.Sprintf("kind           %s", nodeKind(node))},
		{Kind: lineText, Text: fmt.Sprintf("path           %s", node.AbsPath)},
	}

	if node.IsDir {
		items = append(items,
			inspectLine{Kind: lineText, Text: fmt.Sprintf("children       %d", len(node.Children))},
			inspectLine{Kind: lineText, Text: fmt.Sprintf("folders        %d", max(0, stats.Dirs-1))},
			inspectLine{Kind: lineText, Text: fmt.Sprintf("files          %d", stats.Files)},
			inspectLine{Kind: lineText, Text: fmt.Sprintf("state          %s", expandedState(node))},
		)
	} else {
		items = append(items,
			inspectLine{Kind: lineText, Text: fmt.Sprintf("extension      %s", emptyFallback(node.Ext, "none"))},
			inspectLine{Kind: lineText, Text: fmt.Sprintf("size           %s", formatBytes(node.Size))},
		)
	}

	items = append(items,
		inspectLine{Kind: lineBlank},
		inspectLine{Kind: lineSection, Text: "Preview"},
	)

	if strings.TrimSpace(m.Preview.Notice) != "" {
		items = append(items, inspectLine{Kind: lineWarning, Text: m.Preview.Notice})
		if len(m.Preview.Lines) > 0 {
			items = append(items, inspectLine{Kind: lineBlank})
		}
	}

	if len(m.Preview.Lines) == 0 {
		items = append(items, inspectLine{Kind: lineMuted, Text: "No preview available."})
	} else if m.Preview.Numbered {
		for i, line := range m.Preview.Lines {
			items = append(items, inspectLine{
				Kind:   linePreviewNumbered,
				Number: i + 1,
				Text:   line,
			})
		}
	} else {
		for _, line := range m.Preview.Lines {
			items = append(items, inspectLine{
				Kind: linePreview,
				Text: line,
			})
		}
	}

	return items
}

func (m Model) renderInspectLines(th theme.Theme, width, height int) []string {
	items := m.buildInspectLineItems()
	if len(items) == 0 {
		return []string{th.Muted.Render("No inspect content.")}
	}

	offset := clamp(m.InspectScrollOffset, 0, max(0, len(items)-height))
	end := min(len(items), offset+height)

	lines := make([]string, 0, end-offset)
	for _, item := range items[offset:end] {
		lines = append(lines, renderInspectLine(th, item, width))
	}

	if len(lines) == 0 {
		return []string{th.Muted.Render("No inspect content.")}
	}

	return lines
}

func renderInspectLine(th theme.Theme, item inspectLine, width int) string {
	switch item.Kind {
	case lineBlank:
		return ""

	case lineTitle:
		return th.Text.Render(truncateRunes(item.Text, width))

	case lineSubtitle:
		return th.Muted.Render(truncateMiddleRunes(item.Text, width))

	case lineSection:
		return th.Accent.Render(truncateRunes(item.Text, width))

	case lineText:
		return th.Text.Render("  " + truncateRunes(item.Text, max(0, width-2)))

	case lineMuted:
		return th.Muted.Render("  " + truncateRunes(item.Text, max(0, width-2)))

	case lineWarning:
		return th.Warning.Render("  " + truncateRunes(item.Text, max(0, width-2)))

	case linePreview:
		return th.Text.Render("  " + truncateRunes(item.Text, max(0, width-2)))

	case linePreviewNumbered:
		numberPrefix := fmt.Sprintf("%2d │", item.Number)
		textWidth := max(1, width-lipgloss.Width(numberPrefix)-1)
		return th.Meta.Render(numberPrefix) + " " + th.Text.Render(truncateRunes(item.Text, textWidth))
	}

	return truncateRunes(item.Text, width)
}

func (m Model) renderFooter(th theme.Theme) string {
	maxWidth := max(20, m.Width-th.App.GetHorizontalFrameSize())

	if m.Loading {
		controls := footerLabel(th, "STARTUP") + "  " + footerCmd(th, "q", "uit")
		status := footerLabel(th, "STATUS") + "  " + th.Status.Render("startup  ·  "+m.StatusMessage)

		return th.Footer.Render(
			"\n" +
				footerFit(maxWidth, controls) + "\n" +
				footerFit(maxWidth, status),
		)
	}

	if m.LoadError != "" {
		controls := footerLabel(th, "RECOVERY") + "  " + strings.Join([]string{
			footerCmd(th, "r", "etry"),
			footerCmd(th, "q", "uit"),
		}, "   ")
		status := footerLabel(th, "STATUS") + "  " + th.Status.Render("startup  ·  "+m.StatusMessage)

		return th.Footer.Render(
			"\n" +
				footerFit(maxWidth, controls) + "\n" +
				footerFit(maxWidth, status),
		)
	}

	if m.isCompactLayout() {
		return m.renderCompactFooter(th)
	}

	actions := footerLabel(th, "ACTIONS") + "  " + strings.Join([]string{
		footerCmd(th, "q", "uit"),
		footerCmd(th, "t", "oggle"),
		footerCmd(th, "s", "ave/export"),
		footerCmd(th, "c", "lipboard"),
		footerCmd(th, "f", "ilters"),
		footerCmd(th, "w", "orkspace"),
		footerCmd(th, "e", "xpand/collapse"),
	}, "   ")

	navigation := footerLabel(th, "NAV") + "      " + strings.Join([]string{
		footerCmd(th, "tab", " pane"),
		footerCmd(th, "↑↓", " move/scroll"),
		footerCmd(th, "←→", " tree"),
		footerCmd(th, "PgUp/PgDn", " page"),
		footerCmd(th, "Home/End", " bounds"),
		footerCmd(th, "/", " search"),
		footerCmd(th, "?", " help"),
		footerCmd(th, "click", " focus/select"),
		footerCmd(th, "wheel", " hover-scroll"),
	}, "   ")

	return th.Footer.Render(
		"\n" +
			footerFit(maxWidth, actions) + "\n" +
			footerFit(maxWidth, navigation),
	)
}

func (m Model) renderHelpView(th theme.Theme) string {
	lines := []string{
		th.HeaderTitle.Render("copytool / help"),
		"",
		th.Text.Render("B15B is live: the workspace drawer now has an internal file/folder picker for root and output path selection."),
		th.Muted.Render("This keeps the workflow fully inside the TUI while avoiding platform-specific native dialogs."),
		"",
		th.Accent.Render("Workspace drawer"),
		"  • w opens the workspace drawer",
		"  • Ctrl+P opens the picker for the focused workspace field",
		"  • root field uses folder-picking behavior",
		"  • output field allows directory or file target selection",
		"",
		th.Accent.Render("Picker controls"),
		"  • ↑ / ↓ = move",
		"  • Enter = open directory / choose file / choose current directory row",
		"  • Space = choose highlighted path directly",
		"  • Backspace / ← = go to parent directory",
		"  • Esc = close picker",
		"",
		th.Accent.Render("Keyboard"),
		fmt.Sprintf("  %s  switch active pane", th.Key.Render("tab")),
		fmt.Sprintf("  %s  move upward / scroll upward", th.Key.Render("↑ / k")),
		fmt.Sprintf("  %s  move downward / scroll downward", th.Key.Render("↓ / j")),
		fmt.Sprintf("  %s  page upward", th.Key.Render("PgUp / Ctrl+U")),
		fmt.Sprintf("  %s  page downward", th.Key.Render("PgDn / Ctrl+D")),
		fmt.Sprintf("  %s  top / bottom", th.Key.Render("Home / End")),
		fmt.Sprintf("  %s  expand focused folder", th.Key.Render("→ / l")),
		fmt.Sprintf("  %s  collapse focused folder / jump to parent", th.Key.Render("← / h")),
		fmt.Sprintf("  %s  toggle focused folder open/closed", th.Key.Render("e")),
		fmt.Sprintf("  %s  toggle selection", th.Key.Render("t / space")),
		fmt.Sprintf("  %s  open filters drawer", th.Key.Render("f")),
		fmt.Sprintf("  %s  open workspace drawer", th.Key.Render("w")),
		fmt.Sprintf("  %s  export selection", th.Key.Render("s")),
		fmt.Sprintf("  %s  toggle clipboard mode", th.Key.Render("c")),
		fmt.Sprintf("  %s  quit / trigger confirmation if dirty", th.Key.Render("q")),
		fmt.Sprintf("  %s  toggle this help screen", th.Key.Render("?")),
		"",
		th.Accent.Render("Next"),
		"  • command/search palette",
		"",
		th.Muted.Render("Press ? again to return."),
	}

	width := min(100, max(56, m.Width-10))

	panel := th.PanelFocused.
		Width(width).
		Render(strings.Join(lines, "\n"))

	return placeOverlay(m, max(40, m.Width-2), max(12, m.Height-2), panel)
}

func (m Model) renderQuitConfirmView(th theme.Theme) string {
	selected := m.selectedFilesCount()
	width := min(84, max(52, m.Width-12))
	actionsRow := renderQuitConfirmActionsRow(th, quitConfirmContentWidth(th.PanelFocused, width))

	lines := []string{
		th.HeaderTitle.Render("Quit without exporting?"),
		"",
		th.Warning.Render("You have unexported selection changes."),
		"",
		th.Text.Render(fmt.Sprintf("selected files  %d", selected)),
		th.Text.Render(fmt.Sprintf("root            %s", m.RootPath)),
		"",
		th.Accent.Render("Actions"),
		actionsRow,
	}

	panel := th.PanelFocused.
		Width(width).
		Render(strings.Join(lines, "\n"))

	return placeOverlay(m, max(40, m.Width-2), max(12, m.Height-2), panel)
}

func (m Model) renderSaveSuccessView(th theme.Theme) string {
	width := min(88, max(56, m.Width-12))
	contentWidth := quitConfirmContentWidth(th.PanelFocused, width)
	actionsRow := renderSaveSuccessActionsRow(th, contentWidth)

	output := emptyFallback(m.ResolvedOutputPath, "tools/output.txt")
	status := truncateRunes(m.StatusMessage, max(20, contentWidth-2))

	lines := []string{
		th.HeaderTitle.Render("Export complete"),
		"",
		th.Success.Render("Selection saved successfully."),
		"",
		th.Text.Render(fmt.Sprintf("selected files  %d", m.selectedFilesCount())),
		th.Text.Render("output          " + truncateMiddleRunes(output, max(18, contentWidth-16))),
		th.Muted.Render(status),
		"",
		th.Accent.Render("Next"),
		actionsRow,
	}

	panel := th.PanelFocused.Copy().
		Width(width).
		Render(strings.Join(lines, "\n"))

	return placeOverlay(m, max(40, m.Width-2), max(12, m.Height-2), panel)
}

func renderQuitConfirmActionsRow(th theme.Theme, contentWidth int) string {
	quitBtn := th.QuitConfirmKey.Render("[q]") + "uit without saving"
	saveBtn := th.QuitConfirmKey.Render("[s]") + "ave/export & stay"
	cancelBtn := th.QuitConfirmKey.Render("[Esc]/[n]/[c]") + "ancel"

	return justifyThree(contentWidth, quitBtn, saveBtn, cancelBtn)
}

func renderSaveSuccessActionsRow(th theme.Theme, contentWidth int) string {
	continueBtn := th.QuitConfirmKey.Render("[c]") + "ontinue"
	quitBtn := th.QuitConfirmKey.Render("[q]") + "uit"

	return justifyTwo(contentWidth, continueBtn, quitBtn)
}

func justifyTwo(width int, left, right string) string {
	if width <= 0 {
		return lipgloss.JoinHorizontal(lipgloss.Left, left, " ", right)
	}

	leftW := lipgloss.Width(left)
	rightW := lipgloss.Width(right)
	used := leftW + rightW

	if used >= width {
		return lipgloss.JoinHorizontal(lipgloss.Left, left, " ", right)
	}

	gap := max(1, width-used)
	return left + strings.Repeat(" ", gap) + right
}

func justifyThree(width int, left, middle, right string) string {
	if width <= 0 {
		return lipgloss.JoinHorizontal(lipgloss.Left, left, " ", middle, " ", right)
	}

	leftW := lipgloss.Width(left)
	middleW := lipgloss.Width(middle)
	rightW := lipgloss.Width(right)

	used := leftW + middleW + rightW
	if used >= width {
		return lipgloss.JoinHorizontal(lipgloss.Left, left, " ", middle, " ", right)
	}

	gap := width - used
	gapLeft := max(1, gap/2)
	gapRight := max(1, gap-gapLeft)

	return left + strings.Repeat(" ", gapLeft) + middle + strings.Repeat(" ", gapRight) + right
}

func quitConfirmContentWidth(style lipgloss.Style, panelWidth int) int {
	w := panelWidth - style.GetHorizontalFrameSize()
	if w < 1 {
		return 1
	}
	return w
}

func placeOverlay(m Model, width, height int, content string) string {
	return lipgloss.Place(
		width,
		height,
		lipgloss.Center,
		lipgloss.Center,
		content,
	)
}

func renderPanel(style lipgloss.Style, th theme.Theme, title, body string, width, height int) string {
	content := th.PanelTitle.Render(title) + "\n\n" + body
	innerWidth := max(1, width-style.GetHorizontalBorderSize())
	innerHeight := max(1, height-style.GetVerticalBorderSize())

	return style.Copy().
		Width(innerWidth).
		Height(innerHeight).
		Render(content)
}

func summarizePaths(paths []string, limit int) string {
	if len(paths) == 0 {
		return "none"
	}

	if len(paths) <= limit {
		return strings.Join(paths, ", ")
	}

	return strings.Join(paths[:limit], ", ") + fmt.Sprintf(" +%d more", len(paths)-limit)
}

func windowRange(total, focus, size int) (int, int) {
	if total <= 0 {
		return 0, 0
	}

	if size >= total {
		return 0, total
	}

	start := focus - size/2
	if start < 0 {
		start = 0
	}

	end := start + size
	if end > total {
		end = total
		start = end - size
	}

	if start < 0 {
		start = 0
	}

	return start, end
}

func nodeKind(node *core.Node) string {
	if node == nil {
		return "unknown"
	}
	if node.IsDir {
		return "folder"
	}
	return "file"
}

func expandedState(node *core.Node) string {
	if node == nil || !node.IsDir {
		return "n/a"
	}
	if node.Expanded {
		return "expanded"
	}
	return "collapsed"
}

func checkboxForNode(node *core.Node) string {
	if node == nil {
		return "[ ]"
	}

	if node.Partial {
		return "[-]"
	}

	if node.Selected {
		return "[x]"
	}

	return "[ ]"
}

func formatBytes(size int64) string {
	const unit = 1024
	if size < unit {
		return fmt.Sprintf("%d B", size)
	}

	div, exp := int64(unit), 0
	for n := size / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}

	return fmt.Sprintf("%.1f %ciB", float64(size)/float64(div), "KMGTPE"[exp])
}

func truncateRunes(s string, width int) string {
	if width <= 0 {
		return ""
	}

	runes := []rune(s)
	if len(runes) <= width {
		return s
	}

	if width == 1 {
		return "…"
	}

	return string(runes[:width-1]) + "…"
}

func truncateMiddleRunes(s string, width int) string {
	if width <= 0 {
		return ""
	}

	runes := []rune(s)
	if len(runes) <= width {
		return s
	}

	if width <= 3 {
		return truncateRunes(s, width)
	}

	left := (width - 1) / 2
	right := width - left - 1

	return string(runes[:left]) + "…" + string(runes[len(runes)-right:])
}

func emptyFallback(value, fallback string) string {
	if strings.TrimSpace(value) == "" {
		return fallback
	}
	return value
}

func panelContentHeight(panelHeight int) int {
	h := panelHeight - 6
	if h < 1 {
		return 1
	}
	return h
}

func panelContentWidth(panelWidth int) int {
	w := panelWidth - 6
	if w < 8 {
		return 8
	}
	return w
}

func (m Model) activePaneLabel() string {
	if m.ActivePane == paneInspect {
		return "inspect"
	}
	return "navigator"
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func keyHint(th theme.Theme, key, desc string) string {
	return th.Key.Render(key) + " " + th.Muted.Render(desc)
}

func keyHintQC(th theme.Theme, key, desc string) string {
	return th.QuitConfirmKey.Render(key) + th.Muted.Render(desc)
}

func footerLabel(th theme.Theme, label string) string {
	return th.Accent.Copy().Bold(true).Render("::" + label)
}

func footerCmd(th theme.Theme, key, desc string) string {
	keyPart := th.Text.Bold(true).Background(lipgloss.Color("#000000")).Render("[" + key + "]")
	textPart := th.Accent.Background(lipgloss.Color("#2f3132")).Render(desc)
	return keyPart + textPart
}

func footerRule(th theme.Theme, width int) string {
	if width <= 0 {
		return ""
	}
	return th.Meta.Render(strings.Repeat("─", width))
}

func footerFit(width int, line string) string {
	if width <= 0 {
		return line
	}
	return lipgloss.NewStyle().MaxWidth(width).Render(line)
}

func joinHintRow(items []string) string {
	return strings.Join(items, "   ")
}
