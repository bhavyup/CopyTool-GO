package tui

import (
	"bytes"
	"fmt"
	"path/filepath"
	"strings"

	"copytool/internal/core"
	"copytool/internal/theme"

	"github.com/alecthomas/chroma/v2/lexers"
	"github.com/alecthomas/chroma/v2/quick"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
)

type inspectLineKind int

const (
	lineBlank inspectLineKind = iota
	lineTitle
	lineSubtitle
	lineSection
	lineKeyVal
	lineText
	lineMuted
	lineWarning
	linePreview
	linePreviewNumbered
)

type inspectLine struct {
	Kind   inspectLineKind
	Key    string
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
	headerStyle := th.HeaderBar.Copy()
	headerContentWidth := max(1, appContentWidth-headerStyle.GetHorizontalFrameSize())
	detailWidth := max(18, headerContentWidth-10)

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

	clipboardLabel := "[c]lipboard: off"
	if m.ClipboardEnabled {
		clipboardLabel = "[c]lipboard: on"
	}

	selectedStyle := th.HeaderStat
	if selectedFiles > 0 {
		selectedStyle = th.HeaderStat.Copy().Foreground(th.P.SelFull).Bold(true)
	}

	savedChip := th.HeaderBadge.Copy().
		Background(lipgloss.Color("#1A2E1A")).
		Foreground(th.P.Success).
		Bold(true).
		Render("✓ saved")

	stateItems := []string{
		th.HeaderBadge.Render(m.filterBadgeLabel()),
		th.HeaderBadge.Render(clipboardLabel),
	}

	if m.Dirty {
		stateItems = append(stateItems, th.HeaderDirty.Render("● draft"))
	} else {
		stateItems = append(stateItems, savedChip)
	}

	if m.Loading {
		stateItems = append(stateItems, th.HeaderBadge.Render("scanning"))
	} else if m.LoadError != "" {
		stateItems = append(stateItems, th.HeaderDirty.Copy().Background(th.P.Danger).Render("scan failed"))
	}

	dot := th.Muted.Background(th.P.Surface).Render(" · ")
	leftGroup := lipgloss.JoinHorizontal(
		lipgloss.Left,
		th.HeaderBrand.Render("copytool"),
		th.TextWS.Render(" "),
		th.HeaderProject.Render(project),
		th.TextWS.Render(" ▏"),
		selectedStyle.Render(fmt.Sprintf("%d selected", selectedFiles)),
		dot,
		th.HeaderStat.Render(fmt.Sprintf("%d files", stats.Files)),
		dot,
		th.HeaderStat.Render(fmt.Sprintf("%d dirs", stats.Dirs)),
		dot,
		th.HeaderStat.Render(fmt.Sprintf("%d visible", visibleCount)),
	)

	rightGroup := strings.Join(stateItems, th.TextWS.Render(" "))
	titleRow := padBetween(th, leftGroup, rightGroup, headerContentWidth)
	rule := th.Muted.Background(th.P.Surface).Render(strings.Repeat("─", headerContentWidth))

	details := []string{
		renderHeaderDetailRow(th, "root", truncateMiddleRunes(m.RootPath, detailWidth)),
		renderHeaderDetailRow(th, "output", truncateMiddleRunes(emptyFallback(m.ResolvedOutputPath, "none"), detailWidth)),
		renderHeaderDetailRow(th, "focus", truncateMiddleRunes(emptyFallback(focusedPath, "none"), detailWidth)),
	}

	if len(m.InitialSelections) > 0 {
		details = append(details,
			renderHeaderDetailRow(th, "seed", truncateRunes(summarizePaths(m.InitialSelections, 3), detailWidth)),
		)
	}

	content := titleRow + "\n" + rule + "\n" + strings.Join(details, "\n")

	rendered := headerStyle.
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

func renderHeaderDetailRow(th theme.Theme, key, value string) string {
	return th.InspectVal.Render(padRight(key, 8)) + th.TextWS.Render(" ") + th.InspectKey.Background(lipgloss.Color("#000000")).Render(value)
}

func (m Model) renderBody(th theme.Theme, height int) string {
	availableWidth := max(40, m.Width-th.App.GetHorizontalFrameSize())
	gap := 1

	leftWidth := (availableWidth - gap) * 54 / 100
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
	titleStyle := th.Muted
	if m.ActivePane == paneTree {
		style = th.PanelFocused
		titleStyle = th.InspectSection.Copy().Bold(true)
	}

	activeIndicator := ""
	if m.ActivePane == paneTree {
		activeIndicator = "●"
	}

	titleRow := padBetween(th, titleStyle.Render("Navigator"), th.StatusSuccess.Render(activeIndicator), contentWidth)
	rule := titleStyle.Background(th.P.Surface).Render(strings.Repeat("─", contentWidth))
	content := titleRow + "\n" + rule + "\n" + body

	innerWidth := max(1, width-style.GetHorizontalBorderSize())
	innerHeight := max(1, height-style.GetVerticalBorderSize())

	return style.Copy().
		Width(innerWidth).
		Height(innerHeight).
		Render(content)
}

func (m Model) renderInspectPanel(th theme.Theme, width, height int) string {
	contentWidth := panelContentWidth(width)
	contentHeight := panelContentHeight(height)

	bodyLines := m.renderInspectLines(th, contentWidth, contentHeight)
	body := strings.Join(bodyLines, "\n")

	style := th.Panel
	titleStyle := th.Muted

	if m.ActivePane == paneInspect {
		style = th.PanelFocused
		titleStyle = th.InspectSection.Copy().Bold(true)
	}

	activeIndicator := ""
	if m.ActivePane == paneInspect {
		activeIndicator = "●"
	}

	titleRow := titleStyle.Render("Inspect")
	titleRow = padBetween(th, titleStyle.Render("Inspect"), th.StatusSuccess.Render(activeIndicator), contentWidth)

	rule := titleStyle.Background(th.P.Surface).Render(strings.Repeat("─", contentWidth))
	content := titleRow + "\n" + rule + "\n" + body

	innerWidth := max(1, width-style.GetHorizontalBorderSize())
	innerHeight := max(1, height-style.GetVerticalBorderSize())

	return style.Copy().
		Width(innerWidth).
		Height(innerHeight).
		Render(content)
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
	indent := buildIndent(th, depth, focused)

	bg := th.P.Raised
	TreeIndent := th.TreeIndent
	TreeRowFile := th.TreeRowFile
	TreeRowDir := th.TreeRowDir
	TreeRowExt := th.TreeRowExt
	TextWS := th.TextWS

	if focused {
		TreeIndent = TreeIndent.Copy().Background(bg)
		TreeRowFile = TreeRowFile.Copy().Background(bg)
		TreeRowDir = TreeRowDir.Copy().Background(bg)
		TreeRowExt = TreeRowExt.Copy().Background(bg)
		TextWS = TextWS.Copy().Background(bg)
	}

	arrow := TreeIndent.Render("●")
	if node.IsDir {
		arrow = theme.DirArrow(th, node.Expanded, len(node.Children) > 0, focused)
	}

	selGlyph := theme.SelectionGlyph(th, node.Selected, node.Partial, focused)

	nameStr := TreeRowFile.Render(node.Name)
	if node.IsDir {
		nameStr = TreeRowDir.Render(node.Name + "/")
	}

	metaTag := ""
	if node.IsDir {
		sel := core.CountSelectedFiles(node)
		metaTag = TreeRowExt.Render(fmt.Sprintf("%d children · %d selected", len(node.Children), sel))
	} else if node.Ext != "" {
		metaTag = TreeRowExt.Render(strings.TrimPrefix(node.Ext, "."))
	}

	prefixRaw := indent + arrow + TextWS.Render(" ") + selGlyph + TextWS.Render(" ")
	prefixWidth := lipgloss.Width(prefixRaw)
	metaWidth := lipgloss.Width(metaTag)

	rowWidth := width
	if focused {
		rowWidth = max(1, width-th.RowFocused.GetHorizontalFrameSize())
	}

	nameMaxWidth := rowWidth - prefixWidth - metaWidth - 1
	if nameMaxWidth < 1 {
		nameMaxWidth = 1
	}

	if lipgloss.Width(nameStr) > nameMaxWidth {
		nameStr = truncateStyledName(th, node.Name, node.IsDir, nameMaxWidth)
	}

	gap := rowWidth - prefixWidth - lipgloss.Width(nameStr) - metaWidth
	if gap < 1 {
		gap = 1
	}

	row := prefixRaw + nameStr + strings.Repeat(TextWS.Render(" "), gap) + metaTag

	if focused {
		return th.RowFocused.Render(row)
	}

	return row
}

func buildIndent(th theme.Theme, depth int, focused bool) string {
	if depth == 0 {
		return ""
	}

	TreeIndent := th.TreeIndent
	bg := th.P.Raised

	if focused {
		TreeIndent = TreeIndent.Copy().Background(bg)
	}

	var b strings.Builder
	for i := 0; i < depth; i++ {
		b.WriteString(TreeIndent.Render("  "))
	}

	return b.String()
}

func truncateStyledName(th theme.Theme, rawName string, isDir bool, maxWidth int) string {
	suffix := ""
	if isDir {
		suffix = "/"
	}

	runes := []rune(rawName + suffix)
	if maxWidth <= 1 {
		if isDir {
			return th.TreeRowDir.Render("…")
		}
		return th.TreeRowFile.Render("…")
	}

	if len(runes) > maxWidth-1 {
		runes = runes[:maxWidth-1]
	}
	truncated := string(runes) + "…"

	if isDir {
		return th.TreeRowDir.Render(truncated)
	}
	return th.TreeRowFile.Render(truncated)
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
			{Kind: lineMuted, Text: "tree not loaded"},
		}
	}

	node := m.Tree.FocusedNode()
	if node == nil {
		return []inspectLine{
			{Kind: lineMuted, Text: "nothing focused"},
		}
	}

	stats := core.ComputeStats(node)
	selectedFilesInNode := core.CountSelectedFiles(node)
	items := []inspectLine{
		{Kind: lineTitle, Text: node.Name},
		{Kind: lineSubtitle, Text: node.RelPath},
		{Kind: lineBlank},
		{Kind: lineSection, Text: "SELECTION"},
		{Kind: lineKeyVal, Key: "state", Text: core.SelectionState(node)},
		{Kind: lineKeyVal, Key: "files", Text: fmt.Sprintf("%d", selectedFilesInNode)},
		{Kind: lineBlank},
		{Kind: lineSection, Text: "METADATA"},
		{Kind: lineKeyVal, Key: "kind", Text: nodeKind(node)},
		{Kind: lineKeyVal, Key: "path", Text: node.AbsPath},
	}

	if node.IsDir {
		items = append(items,
			inspectLine{Kind: lineKeyVal, Key: "children", Text: fmt.Sprintf("%d", len(node.Children))},
			inspectLine{Kind: lineKeyVal, Key: "folders", Text: fmt.Sprintf("%d", max(0, stats.Dirs-1))},
			inspectLine{Kind: lineKeyVal, Key: "total files", Text: fmt.Sprintf("%d", stats.Files)},
			inspectLine{Kind: lineKeyVal, Key: "expanded", Text: expandedState(node)},
		)
	} else {
		items = append(items,
			inspectLine{Kind: lineKeyVal, Key: "extension", Text: emptyFallback(node.Ext, "none")},
			inspectLine{Kind: lineKeyVal, Key: "size", Text: formatBytes(node.Size)},
		)
	}

	items = append(items,
		inspectLine{Kind: lineBlank},
		inspectLine{Kind: lineSection, Text: "PREVIEW"},
	)

	if strings.TrimSpace(m.Preview.Notice) != "" {
		items = append(items, inspectLine{Kind: lineWarning, Text: m.Preview.Notice})
	}

	if len(m.Preview.Lines) == 0 {
		items = append(items, inspectLine{Kind: lineMuted, Text: "no preview available"})
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
		return []string{th.InspectMuted.Render("no content")}
	}

	lexer := m.previewLexerAlias()

	offset := clamp(m.InspectScrollOffset, 0, max(0, len(items)-height))
	end := min(len(items), offset+height)

	lines := make([]string, 0, end-offset)
	for _, item := range items[offset:end] {
		lines = append(lines, renderInspectLine(th, item, width, lexer))
	}

	if len(lines) == 0 {
		return []string{th.InspectMuted.Render("no content")}
	}

	return lines
}

func (m Model) previewLexerAlias() string {
	if m.Tree == nil {
		return ""
	}

	node := m.Tree.FocusedNode()
	if node == nil || node.IsDir {
		return ""
	}

	lexer := lexers.Match(node.Name)
	if lexer == nil {
		lexer = lexers.Match(node.AbsPath)
	}
	if lexer == nil {
		return ""
	}

	cfg := lexer.Config()
	if cfg == nil {
		return ""
	}
	if len(cfg.Aliases) > 0 {
		return cfg.Aliases[0]
	}

	return strings.ToLower(cfg.Name)
}

func renderInspectLine(th theme.Theme, item inspectLine, width int, lexer string) string {
	switch item.Kind {
	case lineBlank:
		return ""

	case lineTitle:
		return th.InspectTitle.Render(truncateRunes(item.Text, width))

	case lineSubtitle:
		return th.InspectSubtitle.Render(truncateMiddleRunes(item.Text, width))

	case lineSection:
		return theme.RuledSection(th, item.Text, width)

	case lineKeyVal:
		const keyCol = 13
		valWidth := max(1, width-keyCol)
		key := padRight(item.Key, keyCol)
		val := truncateMiddleRunes(item.Text, valWidth)
		return th.InspectKey.Render(key) + th.InspectVal.Render(val)

	case lineText:
		return th.InspectVal.Render("  " + truncateRunes(item.Text, max(0, width-2)))

	case lineMuted:
		return th.InspectMuted.Render("  " + truncateRunes(item.Text, max(0, width-2)))

	case lineWarning:
		return th.InspectWarning.Render("  ⚠  " + truncateRunes(item.Text, max(0, width-5)))

	case linePreview:
		return renderInspectPreviewLine(th, item.Text, width)

	case linePreviewNumbered:
		gutter := fmt.Sprintf("%4d", item.Number)
		gutterStr := th.InspectLineNum.Render(gutter + " ╎ ")
		gutterW := lipgloss.Width(gutterStr)
		codeW := max(1, width-gutterW)
		return gutterStr + renderHighlightedPreviewCode(th, item.Text, codeW, lexer)
	}

	return truncateRunes(item.Text, width)
}

func renderHighlightedPreviewCode(th theme.Theme, line string, width int, lexer string) string {
	if width <= 0 {
		return ""
	}

	fallback := th.InspectCode.Render(truncateRunes(line, width))
	if lexer == "" {
		return fallback
	}

	var b bytes.Buffer
	if err := quick.Highlight(&b, line, lexer, "terminal16m", "github-dark"); err != nil {
		return fallback
	}

	highlighted := strings.TrimSuffix(b.String(), "\n")
	if highlighted == "" {
		return ""
	}

	return ansi.Truncate(highlighted, width, "…")
}

func renderInspectPreviewLine(th theme.Theme, line string, width int) string {
	trimmed := strings.TrimSpace(line)
	if trimmed == "" {
		return ""
	}

	if strings.EqualFold(trimmed, "subtree snapshot") || strings.EqualFold(trimmed, "sample contents") {
		return th.InspectSection.Render("  " + strings.ToUpper(trimmed))
	}

	if key, val, ok := parsePreviewKeyValue(trimmed); ok {
		const keyCol = 16
		valWidth := max(1, width-2-keyCol)
		return "  " + th.InspectKey.Render(padRight(key, keyCol)) + th.InspectVal.Render(truncateMiddleRunes(val, valWidth))
	}

	if row, ok := renderPreviewSampleItem(th, trimmed, width); ok {
		return row
	}

	if strings.HasPrefix(trimmed, "…") {
		return th.InspectMuted.Render("  " + truncateRunes(trimmed, max(0, width-2)))
	}

	return th.InspectCode.Render("  " + truncateRunes(line, max(0, width-2)))
}

func parsePreviewKeyValue(line string) (string, string, bool) {
	idx := strings.Index(line, ":")
	if idx <= 0 {
		return "", "", false
	}

	key := strings.TrimSpace(line[:idx])
	val := strings.TrimSpace(line[idx+1:])
	if key == "" || val == "" {
		return "", "", false
	}

	return key + ":", val, true
}

func renderPreviewSampleItem(th theme.Theme, line string, width int) (string, bool) {
	prefix := ""
	rest := line

	switch {
	case strings.HasPrefix(rest, "▸ "):
		prefix = "▸"
		rest = strings.TrimPrefix(rest, "▸ ")
	case strings.HasPrefix(rest, "· "):
		prefix = "·"
		rest = strings.TrimPrefix(rest, "· ")
	default:
		return "", false
	}

	check := ""
	switch {
	case strings.HasPrefix(rest, "[x] "):
		check = "[x]"
		rest = strings.TrimPrefix(rest, "[x] ")
	case strings.HasPrefix(rest, "[-] "):
		check = "[-]"
		rest = strings.TrimPrefix(rest, "[-] ")
	case strings.HasPrefix(rest, "[ ] "):
		check = "[ ]"
		rest = strings.TrimPrefix(rest, "[ ] ")
	default:
		return "", false
	}

	name := strings.TrimSpace(rest)
	if name == "" {
		return "", false
	}

	var glyph string
	switch check {
	case "[x]":
		glyph = th.TreeSelFull.Render("●")
	case "[-]":
		glyph = th.TreeSelPartial.Render("◐")
	case "[ ]":
		glyph = th.TreeSelNone.Render("○")
	default:
		return "", false
	}

	arrow := th.TreeIndent.Render("·")
	if prefix == "▸" {
		arrow = th.TreeDirClosed.Render("▸")
	}

	isDir := strings.HasSuffix(name, "/")
	nameWidth := max(1, width-8)

	var nameStr string
	if isDir {
		nameStr = truncateStyledName(th, strings.TrimSuffix(name, "/"), true, nameWidth)
	} else {
		nameStr = th.TreeRowFile.Render(truncateRunes(name, nameWidth))
	}

	return th.TextWS.Render("  ") + arrow + th.TextWS.Render(" ") + glyph + th.TextWS.Render(" ") + nameStr, true
}

func (m Model) renderFooter(th theme.Theme) string {
	maxWidth := max(20, m.Width-th.App.GetHorizontalFrameSize())
	rule := footerRule(th, maxWidth)

	if m.Loading {
		rows := []string{
			rule,
			m.footerStatusLine(th, maxWidth, "startup"),
		}
		rows = append(rows, footerHintsRows(th, maxWidth, "commands", []string{
			footerCmd(th, "q", "quit"),
		}, 0)...)

		return th.Footer.Render(strings.Join(rows, "\n"))
	}

	if m.LoadError != "" {
		rows := []string{
			rule,
			m.footerStatusLine(th, maxWidth, "recovery"),
		}
		rows = append(rows, footerHintsRows(th, maxWidth, "commands", []string{
			footerCmdPrimary(th, "r", "retry"),
			footerCmd(th, "q", "quit"),
		}, 0)...)

		return th.Footer.Render(strings.Join(rows, "\n"))
	}

	if m.isCompactLayout() {
		return m.renderCompactFooter(th)
	}

	actions := []string{
		footerCmd(th, "s", "ave/export"),
		footerCmd(th, "q", "uit"),
		footerCmd(th, "t", "oggle"),
		footerCmd(th, "c", "lipboard"),
		footerCmd(th, "f", "ilters"),
		footerCmd(th, "w", "orkspace"),
		footerCmd(th, "e", "xpand/collapse"),
	}

	navigation := []string{
		footerCmd(th, "tab", " pane"),
		footerCmd(th, "↑↓", " move/scroll"),
		footerCmd(th, "←→", " tree"),
		footerCmd(th, "PgUp/PgDn", " page"),
		footerCmd(th, "Home/End", " bounds"),
		footerCmd(th, "/", " search"),
		footerCmd(th, "?", " help"),
		footerCmd(th, "click", " focus/select"),
		footerCmd(th, "wheel", " hover-scroll"),
	}
	labelColWidth := footerLabelColumnWidth(th, "actions", "nav")
	colWidths := footerColumnWidths(actions, navigation)

	rows := []string{
		rule,
		m.footerStatusLine(th, maxWidth, "session"),
		rule,
	}
	rows = append(rows, footerAlignedHintsLine(th, maxWidth, "actions", actions, colWidths, labelColWidth))
	rows = append(rows, footerAlignedHintsLine(th, maxWidth, "nav", navigation, colWidths, labelColWidth))

	return th.Footer.Render(strings.Join(rows, "\n"))
}

func (m Model) renderHelpView(th theme.Theme) string {
	panelStyle := th.PanelFocused.Copy()
	panelW := min(118, max(74, m.Width-10))
	contentW := max(32, panelW-panelStyle.GetHorizontalFrameSize())

	project := filepath.Base(m.RootPath)
	clipChip := th.HeaderBadge.Background(th.P.Surface).Render("clipboard off")
	if m.ClipboardEnabled {
		clipChip = th.HeaderBadge.Copy().Background(th.P.Surface).Foreground(th.P.Success).Bold(true).Render("clipboard on")
	}

	titleRight := strings.Join([]string{
		th.HeaderBadge.Render("project: " + project),
		th.HeaderBadge.Render("pane: " + m.activePaneLabel()),
		th.HeaderBadge.Render(fmt.Sprintf("selected: %d", m.selectedFilesCount())),
		clipChip,
	}, th.TextWS.Render(" ▏"))

	titleRow := padBetween(th, th.TreeRowDir.Render("Help"), titleRight, contentW)
	rule := th.Muted.Background(th.P.Surface).Render(strings.Repeat("─", contentW))

	navTop := []string{
		theme.KeyHintAccent(th, "Tab", " pane"),
		FilterFooterCmd(th, "↑/k", " up"),
		FilterFooterCmd(th, "↓/j", " down"),
		FilterFooterCmd(th, "PgUp/PgDn", " page"),
	}
	navBottom := []string{
		FilterFooterCmd(th, "Home/End", " bounds"),
		FilterFooterCmd(th, "←/h", " collapse"),
		FilterFooterCmd(th, "→/l", " expand"),
		FilterFooterCmd(th, "e", " toggle dir"),
	}
	navRow1, navRow2 := renderFilterControlRows(th, contentW, navTop, navBottom)

	actionTop := []string{
		FilterFooterCmd(th, "t/Space", " toggle"),
		FilterFooterCmd(th, "s", " export"),
		FilterFooterCmd(th, "c", " clipboard"),
		FilterFooterCmd(th, "q", " quit"),
	}
	actionBottom := []string{
		FilterFooterCmd(th, "f", " filters"),
		FilterFooterCmd(th, "w", " workspace"),
		FilterFooterCmd(th, "/", " search"),
		FilterFooterCmd(th, "?", " close help"),
	}
	actionRow1, actionRow2 := renderFilterControlRows(th, contentW, actionTop, actionBottom)

	drawerTop := []string{
		theme.KeyHintAccent(th, "Enter", " apply drawer"),
		FilterFooterCmd(th, "Esc", " close drawer"),
		FilterFooterCmd(th, "Ctrl+R", " reset fields"),
		FilterFooterCmd(th, "Ctrl+P", " browse path"),
	}
	drawerBottom := []string{
		FilterFooterCmd(th, "Ctrl+B", " toggle clipboard"),
		FilterFooterCmd(th, "click", " focus/toggle"),
		FilterFooterCmd(th, "wheel", " pane scroll"),
	}
	drawerRow1, drawerRow2 := renderFilterControlRows(th, contentW, drawerTop, drawerBottom)

	pickerTop := []string{
		FilterFooterCmd(th, "↑/↓", " move"),
		FilterFooterCmd(th, "Enter", " open/select"),
		FilterFooterCmd(th, "Space", " choose path"),
		FilterFooterCmd(th, "Backspace/←", " parent dir"),
	}
	pickerRow1, pickerRow2 := renderFilterControlRows(th, contentW, pickerTop, nil)

	lines := []string{
		titleRow,
		rule,
		th.DrawerHint.Render("Quick map of movement, actions, drawers, and picker behavior."),
		"",
		theme.RuledSection(th, "NAVIGATION", contentW),
		navRow1,
		navRow2,
		"",
		theme.RuledSection(th, "ACTIONS", contentW),
		actionRow1,
		actionRow2,
		"",
		theme.RuledSection(th, "DRAWERS", contentW),
		drawerRow1,
		drawerRow2,
		"",
		theme.RuledSection(th, "PICKER", contentW),
		pickerRow1,
		pickerRow2,
		"",
		th.DrawerHint.Render("Press ? again to return."),
	}

	panel := panelStyle.
		Width(panelW).
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

func padBetween(th theme.Theme, left, right string, width int) string {
	if width <= 0 {
		return left + " " + right
	}

	lw := lipgloss.Width(left)
	rw := lipgloss.Width(right)

	if lw+rw+1 > width {
		if rw >= width {
			return truncateRunes(right, width)
		}
		left = truncateRunes(left, max(1, width-rw-1))
		lw = lipgloss.Width(left)
	}

	gap := width - lw - rw
	if gap < 1 {
		gap = 1
	}

	return left + strings.Repeat(th.Muted.Background(th.P.Surface).Render(" "), gap) + right
}

func padRight(s string, n int) string {
	runes := []rune(s)
	if len(runes) >= n {
		return string(runes[:n])
	}
	return s + strings.Repeat(" ", n-len(runes))
}

func (m Model) footerStatusLine(th theme.Theme, width int, mode string) string {
	statusText := strings.TrimSpace(m.StatusMessage)
	if statusText == "" {
		statusText = "ready"
	}

	parts := []string{footerLabel(th, mode)}
	if mode == "startup" || mode == "recovery" {
		parts = append(parts, th.Muted.Render("initializing"))
	} else {
		parts = append(parts,
			th.Muted.Render("pane")+" "+th.Text.Render(m.activePaneLabel()),
			th.Muted.Render("selected")+" "+th.Text.Render(fmt.Sprintf("%d", m.selectedFilesCount())),
		)

		if m.Tree != nil {
			parts = append(parts, th.Muted.Render("visible")+" "+th.Text.Render(fmt.Sprintf("%d", len(m.Tree.VisibleNodes))))
		}
	}

	left := strings.Join(parts, th.Muted.Render("  •  "))
	// right := th.Status.Render(truncateMiddleRunes(statusText, max(14, width/3)))

	return left + th.Muted.Render("  •  ") + th.Status.Render(statusText)
}

func footerColumnWidth(groups ...[]string) int {
	maxWidth := 0
	for _, group := range groups {
		for _, item := range group {
			w := lipgloss.Width(item)
			if w > maxWidth {
				maxWidth = w
			}
		}
	}

	if maxWidth == 0 {
		return 0
	}

	return maxWidth + 2
}

func footerColumnWidths(groups ...[]string) []int {
	maxCols := 0
	for _, group := range groups {
		if len(group) > maxCols {
			maxCols = len(group)
		}
	}

	if maxCols == 0 {
		return nil
	}

	widths := make([]int, maxCols)
	for col := 0; col < maxCols; col++ {
		w := 0
		for _, group := range groups {
			if col >= len(group) {
				continue
			}
			itemW := lipgloss.Width(group[col])
			if itemW > w {
				w = itemW
			}
		}

		if col < maxCols-1 {
			w += 2
		}

		widths[col] = w
	}

	return widths
}

func footerLabelColumnWidth(th theme.Theme, labels ...string) int {
	maxWidth := 0
	for _, label := range labels {
		w := lipgloss.Width(footerLabel(th, label))
		if w > maxWidth {
			maxWidth = w
		}
	}
	return maxWidth
}

func footerAlignedHintsLine(th theme.Theme, width int, label string, hints []string, colWidths []int, labelColWidth int) string {
	if len(hints) == 0 {
		return ""
	}

	prefix := padStyledExactRight(footerLabel(th, label), labelColWidth) + th.Muted.Render("  ")

	var b strings.Builder
	b.WriteString(prefix)

	for i, hint := range hints {
		if i == len(hints)-1 {
			b.WriteString(hint)
			break
		}

		cellWidth := lipgloss.Width(hint) + 2
		if i < len(colWidths) && colWidths[i] > cellWidth {
			cellWidth = colWidths[i]
		}
		b.WriteString(padStyledRight(hint, cellWidth))
	}

	return ansi.Truncate(b.String(), width, "")
}

func padStyledExactRight(s string, width int) string {
	w := lipgloss.Width(s)
	if w >= width {
		return s
	}

	return s + strings.Repeat(" ", width-w)
}

func footerHintsRows(th theme.Theme, width int, label string, hints []string, cellWidth int) []string {
	if len(hints) == 0 {
		return nil
	}

	prefix := footerLabel(th, label) + th.Muted.Render("  ")
	prefixWidth := lipgloss.Width(prefix)
	available := max(1, width-prefixWidth)

	if cellWidth <= 0 {
		cellWidth = footerColumnWidth(hints)
	}
	if cellWidth < 8 {
		cellWidth = 8
	}

	cols := max(1, available/cellWidth)
	indent := strings.Repeat(" ", prefixWidth)

	rows := make([]string, 0, (len(hints)+cols-1)/cols)
	for i := 0; i < len(hints); i += cols {
		end := min(len(hints), i+cols)
		rowPrefix := indent
		if i == 0 {
			rowPrefix = prefix
		}

		row := rowPrefix + footerColumnLine(hints[i:end], cellWidth)
		rows = append(rows, footerFit(width, row))
	}

	return rows
}

func footerColumnLine(items []string, cellWidth int) string {
	if len(items) == 0 {
		return ""
	}

	if len(items) == 1 {
		return items[0]
	}

	var b strings.Builder
	for i, item := range items {
		if i == len(items)-1 {
			b.WriteString(item)
			break
		}
		b.WriteString(padStyledRight(item, cellWidth))
	}

	return b.String()
}

func padStyledRight(s string, width int) string {
	w := lipgloss.Width(s)
	if w >= width {
		return s + "  "
	}

	return s + strings.Repeat(" ", width-w)
}

func footerLabel(th theme.Theme, label string) string {
	return th.InspectSection.Copy().Bold(true).Render("::" + strings.ToUpper(label))
}

func footerCmd(th theme.Theme, key, desc string) string {
	keyPart := th.TreeRowDir.Bold(true).Background(lipgloss.Color("#000000")).Render("[" + key + "]")
	textPart := th.Status.Render(desc)
	return keyPart + textPart
}

func footerCmdPrimary(th theme.Theme, key, desc string) string {
	return theme.KeyHintAccent(th, key, desc)
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
