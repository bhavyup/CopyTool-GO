package tui

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"copytool/internal/core"
	"copytool/internal/platform"
	"copytool/internal/theme"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type paneFocus string

const (
	paneTree    paneFocus = "tree"
	paneInspect paneFocus = "inspect"
)

type rect struct {
	X int
	Y int
	W int
	H int
}

func (r rect) Contains(x, y int) bool {
	return x >= r.X && x < r.X+r.W && y >= r.Y && y < r.Y+r.H
}

type skeletonTickMsg time.Time

type scanCompleteMsg struct {
	Tree          *core.Tree
	Stats         core.Stats
	Applied       int
	Skipped       int
	SelectedCount int
	Err           error
}

type nativePickResultMsg struct {
	Target    string // "root" or "output"
	Path      string
	Cancelled bool
	Err       error
}

type Model struct {
	RootPath           string
	InitialSelections  []string
	OutputSpec         string
	ResolvedOutputPath string
	Tree               *core.Tree
	Preview            core.PreviewData
	Filters            core.FilterConfig

	Width  int
	Height int

	ActivePane          paneFocus
	InspectScrollOffset int
	ClipboardEnabled    bool
	Dirty               bool

	ShowHelp        bool
	ShowQuitConfirm bool
	ShowSaveSuccess bool
	ShowFilters     bool
	StatusMessage   string
	ShowWorkspace   bool
	ShowPicker      bool

	PickerMode    string
	PickerPath    string
	PickerEntries []pickerEntry
	PickerIndex   int

	WorkspaceRootInput   textinput.Model
	WorkspaceOutputInput textinput.Model
	WorkspaceFocusIndex  int

	Loading               bool
	LoadError             string
	LoadingLabel          string
	SkeletonPhase         int
	PendingScanMode       string
	PendingDirtyAfterScan bool

	FilterIncludeInput textinput.Model
	FilterExcludeInput textinput.Model
	FilterDirsInput    textinput.Model
	FilterFocusIndex   int

	ShowPalette    bool
	PaletteInput   textinput.Model
	PaletteAll     []paletteItem
	PaletteResults []paletteItem
	PaletteIndex   int
}

func NewModel(root string, initialSelections []string, clipboardEnabled bool, filters core.FilterConfig, outputSpec string) (Model, error) {
	absRoot, err := filepath.Abs(root)
	if err != nil {
		return Model{}, fmt.Errorf("resolve root path: %w", err)
	}

	if filters.IncludeExts == nil {
		filters.IncludeExts = map[string]bool{}
	}
	if filters.ExcludeExts == nil {
		filters.ExcludeExts = map[string]bool{}
	}
	if filters.ExcludeDirs == nil {
		filters.ExcludeDirs = core.DefaultUserExcludeDirs()
	}

	resolvedOutputPath, err := core.ResolveOutputPath(absRoot, outputSpec)
	if err != nil {
		return Model{}, fmt.Errorf("resolve output target: %w", err)
	}

	model := Model{
		RootPath:              absRoot,
		InitialSelections:     initialSelections,
		OutputSpec:            outputSpec,
		ResolvedOutputPath:    resolvedOutputPath,
		Filters:               filters,
		ActivePane:            paneTree,
		InspectScrollOffset:   0,
		ClipboardEnabled:      clipboardEnabled,
		Dirty:                 false,
		Loading:               true,
		LoadError:             "",
		LoadingLabel:          "Scanning project workspace...",
		SkeletonPhase:         0,
		PendingScanMode:       "startup",
		PendingDirtyAfterScan: false,
		StatusMessage:         "Scanning project workspace...",
	}

	model.initFilterInputs()
	model.initWorkspaceInputs()
	model.initPalette()

	return model, nil
}

func (m Model) Init() tea.Cmd {
	return tea.Batch(
		skeletonTickCmd(),
		scanProjectCmd(m.RootPath, m.Filters, m.InitialSelections, nil, ""),
	)
}

func skeletonTickCmd() tea.Cmd {
	return tea.Tick(120*time.Millisecond, func(t time.Time) tea.Msg {
		return skeletonTickMsg(t)
	})
}

func scanProjectCmd(
	root string,
	filters core.FilterConfig,
	initialSelections []string,
	selectedRelPaths []string,
	focusedRelPath string,
) tea.Cmd {
	return func() tea.Msg {
		tree, err := core.ScanRoot(root, filters)
		if err != nil {
			return scanCompleteMsg{Err: err}
		}

		stats := core.ComputeStats(tree.Root)
		applied := 0
		skipped := 0

		switch {
		case len(selectedRelPaths) > 0:
			core.ApplyInitialSelections(tree.Root, root, selectedRelPaths)
			tree.RebuildVisible()

		case len(initialSelections) > 0:
			var skippedPaths []string
			applied, skippedPaths = core.ApplyInitialSelections(tree.Root, root, initialSelections)
			skipped = len(skippedPaths)
			tree.RebuildVisible()
		}

		if focusedRelPath != "" {
			if node := core.FindNodeByRelPath(tree.Root, focusedRelPath); node != nil {
				core.ExpandAncestors(node)
				tree.RebuildVisible()
				tree.FocusNode(node)
			}
		}

		selectedCount := 0
		if tree.Root != nil {
			selectedCount = core.CountSelectedFiles(tree.Root)
		}

		return scanCompleteMsg{
			Tree:          tree,
			Stats:         stats,
			Applied:       applied,
			Skipped:       skipped,
			SelectedCount: selectedCount,
			Err:           nil,
		}
	}
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.Width = msg.Width
		m.Height = msg.Height
		m.clampInspectScroll()
		return m, nil

	case skeletonTickMsg:
		if !m.Loading {
			return m, nil
		}

		m.SkeletonPhase = (m.SkeletonPhase + 1) % 12
		return m, skeletonTickCmd()

	case nativePickResultMsg:
		if msg.Err != nil {
			m.StatusMessage = fmt.Sprintf("Picker error: %v", msg.Err)
			return m, nil
		}
		if msg.Cancelled || strings.TrimSpace(msg.Path) == "" {
			m.StatusMessage = "Picker cancelled."
			return m, nil
		}

		if msg.Target == "root" {
			m.WorkspaceRootInput.SetValue(msg.Path)
			m.focusWorkspaceInput(0)
			m.StatusMessage = "Picked root folder."
			return m, nil
		}

		m.WorkspaceOutputInput.SetValue(msg.Path)
		m.focusWorkspaceInput(1)
		m.StatusMessage = "Picked output target."
		return m, nil

	case scanCompleteMsg:
		m.Loading = false

		if msg.Err != nil {
			m.LoadError = msg.Err.Error()
			m.StatusMessage = "Project scan failed. Press r to retry or q to quit."
			return m, nil
		}

		m.LoadError = ""
		m.Tree = msg.Tree
		m.refreshPreview()
		m.rebuildPaletteIndex()

		switch m.PendingScanMode {
		case "startup":
			m.Dirty = msg.SelectedCount > 0

			if len(m.InitialSelections) > 0 {
				m.StatusMessage = fmt.Sprintf(
					"Indexed %d files across %d folders. Applied %d startup selection(s), skipped %d. Selected %d file(s).",
					msg.Stats.Files,
					msg.Stats.Dirs,
					msg.Applied,
					msg.Skipped,
					msg.SelectedCount,
				)
			} else {
				m.StatusMessage = fmt.Sprintf(
					"Indexed %d files across %d folders. Press t to toggle, w for workspace, f for filters, s to export.",
					msg.Stats.Files,
					msg.Stats.Dirs,
				)
			}

		case "workspace":
			m.Dirty = false
			m.ActivePane = paneTree
			m.StatusMessage = fmt.Sprintf(
				"Workspace switched. %d files, %d folders, %d selected file(s).",
				msg.Stats.Files,
				msg.Stats.Dirs,
				msg.SelectedCount,
			)

		default:
			m.Dirty = m.PendingDirtyAfterScan
			m.StatusMessage = fmt.Sprintf(
				"Workspace refreshed. %d files, %d folders, %d selected file(s).",
				msg.Stats.Files,
				msg.Stats.Dirs,
				msg.SelectedCount,
			)
		}

		return m, nil

	case tea.MouseMsg:
		if m.Loading || m.LoadError != "" || m.ShowHelp || m.ShowQuitConfirm || m.ShowSaveSuccess || m.ShowPicker || m.ShowPalette {
			return m, nil
		}

		if m.ShowWorkspace {
			cmd := m.handleWorkspaceMouse(msg)
			return m, cmd
		}

		if m.ShowFilters {
			cmd := m.handleFiltersMouse(msg)
			return m, cmd
		}

		m.handleMouse(msg)
		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			return m, tea.Quit
		}

		if m.Loading {
			switch msg.String() {
			case "q":
				return m, tea.Quit
			}
			return m, nil
		}

		if m.LoadError != "" {
			switch msg.String() {
			case "q":
				return m, tea.Quit
			case "r":
				cmd := m.beginAsyncScan(
					"Retrying project scan...",
					"rescan",
					m.Filters,
					nil,
					m.selectedRelPaths(),
					m.focusedRelPath(),
					m.Dirty,
				)
				return m, cmd
			}
			return m, nil
		}

		if m.ShowQuitConfirm {
			switch msg.String() {
			case "q":
				return m, tea.Quit

			case "esc", "n", "c":
				m.ShowQuitConfirm = false
				m.StatusMessage = "Quit cancelled."
				return m, nil

			case "s":
				m.ShowQuitConfirm = false
				if err := m.performExport(); err != nil {
					m.StatusMessage = fmt.Sprintf("Export failed: %v", err)
				} else {
					m.ShowSaveSuccess = true
				}
				return m, nil
			}

			return m, nil
		}

		if m.ShowSaveSuccess {
			switch msg.String() {
			case "q":
				return m, tea.Quit

			case "c", "enter", "esc":
				m.ShowSaveSuccess = false
				m.StatusMessage = "Export saved. Continue working."
				return m, nil
			}

			return m, nil
		}

		if m.ShowPalette {
			switch msg.String() {
			case "esc":
				m.ShowPalette = false
				m.PaletteInput.Blur()
				m.StatusMessage = "Search closed."
				return m, nil

			case "up", "k":
				if m.PaletteIndex > 0 {
					m.PaletteIndex--
				}
				return m, nil

			case "down", "j":
				if m.PaletteIndex < len(m.PaletteResults)-1 {
					m.PaletteIndex++
				}
				return m, nil

			case "enter":
				m.paletteJumpToHighlighted()
				m.ShowPalette = false
				m.PaletteInput.Blur()
				return m, nil

			case "ctrl+t":
				if len(m.PaletteResults) > 0 && m.Tree != nil && m.Tree.Root != nil {
					item := m.PaletteResults[m.PaletteIndex]
					node := core.FindNodeByRelPath(m.Tree.Root, item.RelPath)
					if node != nil && core.ToggleNode(node) {
						m.Dirty = true
						m.refreshPreview()
						m.StatusMessage = "Toggled: " + item.RelPath
					}
				}
				return m, nil
			}

			var cmd tea.Cmd
			m.PaletteInput, cmd = m.PaletteInput.Update(msg)
			m.paletteRefreshResults()
			return m, cmd
		}

		if m.ShowPicker {
			switch msg.String() {
			case "esc":
				m.closePicker("Picker closed.")
				return m, nil

			case "up", "k":
				m.pickerMove(-1)
				return m, nil

			case "down", "j":
				m.pickerMove(1)
				return m, nil

			case "left", "h", "backspace":
				if err := m.pickerGoParent(); err != nil {
					m.StatusMessage = fmt.Sprintf("Picker error: %v", err)
				}
				return m, nil

			case "enter", "right", "l":
				if err := m.pickerOpenOrSelect(); err != nil {
					m.StatusMessage = fmt.Sprintf("Picker error: %v", err)
				}
				return m, nil

			case " ":
				if err := m.pickerSelectHighlighted(); err != nil {
					m.StatusMessage = fmt.Sprintf("Picker error: %v", err)
				}
				return m, nil
			}

			return m, nil
		}

		if m.ShowWorkspace {
			switch msg.String() {
			case "esc":
				m.closeWorkspace("Workspace drawer closed.")
				return m, nil

			case "tab":
				m.cycleWorkspaceFocus(1)
				return m, textinput.Blink

			case "shift+tab", "backtab":
				m.cycleWorkspaceFocus(-1)
				return m, textinput.Blink

			case "enter":
				cmd, err := m.applyWorkspace()
				if err != nil {
					m.StatusMessage = fmt.Sprintf("Workspace apply failed: %v", err)
					return m, nil
				}
				return m, cmd

			case "ctrl+r":
				m.resetWorkspaceInputsToCurrent()
				m.StatusMessage = "Workspace fields reset to current values."
				return m, textinput.Blink

			case "ctrl+p":
				target := "root"
				if m.WorkspaceFocusIndex == 1 {
					target = "output"
				}
				return m, m.workspaceOpenPickerCmd(target)

			case "ctrl+b":
				m.ClipboardEnabled = !m.ClipboardEnabled
				if m.ClipboardEnabled {
					m.StatusMessage = "Clipboard export enabled."
				} else {
					m.StatusMessage = "Clipboard export disabled."
				}
				return m, nil
			}

			cmd := m.updateFocusedWorkspaceInput(msg)
			return m, cmd
		}

		if m.ShowFilters {
			switch msg.String() {
			case "esc":
				m.closeFilters("Filters closed.")
				return m, nil

			case "tab":
				m.cycleFilterFocus(1)
				return m, textinput.Blink

			case "shift+tab", "backtab":
				m.cycleFilterFocus(-1)
				return m, textinput.Blink

			case "enter":
				cmd := m.beginFilterRescan()
				m.closeFilters("Applying filters in background...")
				return m, cmd

			case "ctrl+r":
				m.resetFilterInputsToDefault()
				m.StatusMessage = "Filter fields reset to defaults."
				return m, textinput.Blink

			case "ctrl+b":
				m.ClipboardEnabled = !m.ClipboardEnabled
				if m.ClipboardEnabled {
					m.StatusMessage = "Clipboard export enabled."
				} else {
					m.StatusMessage = "Clipboard export disabled."
				}
				return m, nil
			}

			cmd := m.updateFocusedFilterInput(msg)
			return m, cmd
		}

		switch msg.String() {
		case "q":
			if m.Dirty {
				m.ShowQuitConfirm = true
				m.StatusMessage = "Unsaved export state. Enter/Y quits, S exports, Esc/N stays."
				return m, nil
			}
			return m, tea.Quit

		case "?":
			m.ShowHelp = !m.ShowHelp
			if m.ShowHelp {
				m.StatusMessage = "Help opened."
			} else {
				m.refreshPreview()
				m.StatusMessage = m.selectionSummaryStatus()
			}
			return m, nil
		}

		if m.ShowHelp {
			return m, nil
		}

		switch msg.String() {
		case "w":
			m.openWorkspace()
			return m, textinput.Blink

		case "f":
			m.openFilters()
			return m, textinput.Blink

		case "tab":
			m.switchActivePane()
			return m, nil

		case "c":
			m.ClipboardEnabled = !m.ClipboardEnabled
			if m.ClipboardEnabled {
				m.StatusMessage = "Clipboard export enabled."
			} else {
				m.StatusMessage = "Clipboard export disabled."
			}
			return m, nil

		case "s":
			if err := m.performExport(); err != nil {
				m.StatusMessage = fmt.Sprintf("Export failed: %v", err)
			} else {
				m.ShowSaveSuccess = true
			}
			return m, nil

		case "up", "k":
			if m.ActivePane == paneInspect {
				if m.scrollInspect(-1) {
					m.StatusMessage = m.inspectStatus()
				}
			} else {
				if m.moveTreeBy(-1) {
					m.StatusMessage = m.focusStatus()
				}
			}
			return m, nil

		case "/":
			m.ShowPalette = true
			m.PaletteInput.SetValue("")
			m.PaletteInput.CursorEnd()
			m.PaletteInput.Focus()
			m.paletteRefreshResults()
			m.StatusMessage = "Search opened."
			return m, textinput.Blink

		case "down", "j":
			if m.ActivePane == paneInspect {
				if m.scrollInspect(1) {
					m.StatusMessage = m.inspectStatus()
				}
			} else {
				if m.moveTreeBy(1) {
					m.StatusMessage = m.focusStatus()
				}
			}
			return m, nil

		case "pgup", "ctrl+u":
			step := max(1, m.inspectViewportHeight()-3)
			if m.ActivePane == paneInspect {
				if m.scrollInspect(-step) {
					m.StatusMessage = m.inspectStatus()
				}
			} else {
				if m.moveTreeBy(-step) {
					m.StatusMessage = m.focusStatus()
				}
			}
			return m, nil

		case "pgdown", "ctrl+d":
			step := max(1, m.inspectViewportHeight()-3)
			if m.ActivePane == paneInspect {
				if m.scrollInspect(step) {
					m.StatusMessage = m.inspectStatus()
				}
			} else {
				if m.moveTreeBy(step) {
					m.StatusMessage = m.focusStatus()
				}
			}
			return m, nil

		case "home":
			if m.ActivePane == paneInspect {
				if m.InspectScrollOffset != 0 {
					m.InspectScrollOffset = 0
					m.StatusMessage = m.inspectStatus()
				}
			} else if m.Tree != nil && len(m.Tree.VisibleNodes) > 0 {
				m.Tree.FocusIndex = 0
				m.refreshPreview()
				m.StatusMessage = m.focusStatus()
			}
			return m, nil

		case "end":
			if m.ActivePane == paneInspect {
				maxOffset := m.inspectMaxScroll()
				if m.InspectScrollOffset != maxOffset {
					m.InspectScrollOffset = maxOffset
					m.StatusMessage = m.inspectStatus()
				}
			} else if m.Tree != nil && len(m.Tree.VisibleNodes) > 0 {
				m.Tree.FocusIndex = len(m.Tree.VisibleNodes) - 1
				m.refreshPreview()
				m.StatusMessage = m.focusStatus()
			}
			return m, nil

		case "e":
			if m.Tree != nil && m.Tree.ToggleFocusedExpand() {
				m.refreshPreview()
				m.StatusMessage = m.focusStatus()
			}
			return m, nil

		case "right", "l":
			if m.ActivePane == paneTree {
				if m.Tree != nil && m.Tree.ExpandFocused() {
					m.refreshPreview()
					m.StatusMessage = m.focusStatus()
				}
			}
			return m, nil

		case "left", "h":
			if m.ActivePane == paneTree {
				if m.Tree != nil && m.Tree.CollapseFocused() {
					m.refreshPreview()
					m.StatusMessage = m.focusStatus()
				}
			}
			return m, nil

		case "t", " ":
			if m.Tree != nil {
				node := m.Tree.FocusedNode()
				if core.ToggleNode(node) {
					m.Dirty = true
					m.refreshPreview()
					m.StatusMessage = m.selectionSummaryStatus()
				}
			}
			return m, nil
		}
	}

	return m, nil
}

func (m *Model) handleWorkspaceMouse(msg tea.MouseMsg) tea.Cmd {
	if msg.Action != tea.MouseActionPress || msg.Button != tea.MouseButtonLeft {
		return nil
	}

	th := theme.New()
	metrics := m.workspaceDrawerSizing(th)
	panelRect, contentX, contentY := m.drawerPanelAndContentRects(th.PanelFocused, metrics.PanelW, metrics.PanelH)
	if !panelRect.Contains(msg.X, msg.Y) {
		return nil
	}
	layout := m.workspaceDrawerLayout(th, metrics)

	clickRow := msg.Y - contentY
	rootStart, rootEnd := renderedLineSpan(layout.Lines, layout.RootLineIndex)
	outStart, outEnd := renderedLineSpan(layout.Lines, layout.OutputLineIndex)
	clipStart, clipEnd := renderedLineSpan(layout.Lines, layout.ClipLineIndex)

	rootBtnX0 := contentX + layout.RootButtonStart
	rootBtnX1 := contentX + layout.RootButtonEnd
	outBtnX0 := contentX + layout.OutputButtonStart
	outBtnX1 := contentX + layout.OutputButtonEnd

	x := msg.X

	// Root line
	if inRenderedSpan(clickRow, rootStart, rootEnd) {
		m.WorkspaceFocusIndex = 0
		m.focusWorkspaceInput(0)

		// If click is inside the pill => open picker
		if x >= rootBtnX0 && x < rootBtnX1 {
			return m.workspaceOpenPickerCmd("root")
		}
		return nil
	}

	// Output line
	if inRenderedSpan(clickRow, outStart, outEnd) {
		m.WorkspaceFocusIndex = 1
		m.focusWorkspaceInput(1)

		if x >= outBtnX0 && x < outBtnX1 {
			return m.workspaceOpenPickerCmd("output")
		}
		return nil
	}

	// Clipboard line toggles anywhere on that line
	if inRenderedSpan(clickRow, clipStart, clipEnd) {
		m.ClipboardEnabled = !m.ClipboardEnabled
		if m.ClipboardEnabled {
			m.StatusMessage = "Clipboard export enabled."
		} else {
			m.StatusMessage = "Clipboard export disabled."
		}
		return nil
	}

	return nil
}

func (m *Model) workspaceOpenPickerCmd(target string) tea.Cmd {
	// Windows: native picker; others: internal picker
	if platform.SupportsNativePicker() {
		return m.nativePickCmd(target)
	}

	// Internal picker: make sure focus index matches target first
	if target == "root" {
		m.WorkspaceFocusIndex = 0
		m.focusWorkspaceInput(0)
	} else {
		m.WorkspaceFocusIndex = 1
		m.focusWorkspaceInput(1)
	}

	if err := m.openPickerForWorkspaceField(); err != nil {
		m.StatusMessage = fmt.Sprintf("Picker open failed: %v", err)
		return nil
	}
	return nil
}

func (m *Model) handleFiltersMouse(msg tea.MouseMsg) tea.Cmd {
	if msg.Action != tea.MouseActionPress || msg.Button != tea.MouseButtonLeft {
		return nil
	}

	th := theme.New()
	metrics := m.filtersDrawerSizing(th)
	panelRect, _, contentY := m.drawerPanelAndContentRects(th.PanelFocused, metrics.PanelW, metrics.PanelH)
	if !panelRect.Contains(msg.X, msg.Y) {
		return nil
	}
	layout := m.filtersDrawerLayout(th, metrics)

	clickRow := msg.Y - contentY
	includeStart, includeEnd := renderedLineSpan(layout.Lines, layout.IncludeLineIndex)
	excludeStart, excludeEnd := renderedLineSpan(layout.Lines, layout.ExcludeLineIndex)
	dirsStart, dirsEnd := renderedLineSpan(layout.Lines, layout.DirsLineIndex)
	clipStart, clipEnd := renderedLineSpan(layout.Lines, layout.ClipLineIndex)

	switch {
	case inRenderedSpan(clickRow, includeStart, includeEnd):
		m.FilterFocusIndex = 0
		m.focusFilterInput(0)
		return nil
	case inRenderedSpan(clickRow, excludeStart, excludeEnd):
		m.FilterFocusIndex = 1
		m.focusFilterInput(1)
		return nil
	case inRenderedSpan(clickRow, dirsStart, dirsEnd):
		m.FilterFocusIndex = 2
		m.focusFilterInput(2)
		return nil
	case inRenderedSpan(clickRow, clipStart, clipEnd):
		m.ClipboardEnabled = !m.ClipboardEnabled
		if m.ClipboardEnabled {
			m.StatusMessage = "Clipboard export enabled."
		} else {
			m.StatusMessage = "Clipboard export disabled."
		}
		return nil
	}

	return nil
}

func renderedLineSpan(lines []string, index int) (int, int) {
	if index < 0 || index >= len(lines) {
		return 0, 0
	}

	start := 0
	for i := 0; i < index; i++ {
		start += renderedLineHeight(lines[i])
	}

	end := start + renderedLineHeight(lines[index])
	return start, end
}

func renderedLineHeight(line string) int {
	h := lipgloss.Height(line)
	if h < 1 {
		return 1
	}
	return h
}

func inRenderedSpan(row, start, end int) bool {
	return row >= start && row < end
}

func (m *Model) beginAsyncScan(
	status string,
	mode string,
	filters core.FilterConfig,
	initialSelections []string,
	selectedRelPaths []string,
	focusedRelPath string,
	dirtyAfter bool,
) tea.Cmd {
	m.Filters = filters
	m.Loading = true
	m.LoadError = ""
	m.LoadingLabel = status
	m.SkeletonPhase = 0
	m.PendingScanMode = mode
	m.PendingDirtyAfterScan = dirtyAfter
	m.StatusMessage = status

	return tea.Batch(
		skeletonTickCmd(),
		scanProjectCmd(m.RootPath, filters, initialSelections, selectedRelPaths, focusedRelPath),
	)
}

func (m *Model) beginFilterRescan() tea.Cmd {
	newFilters := core.FilterConfig{
		IncludeExts: core.ParseExtensionCSV(m.FilterIncludeInput.Value()),
		ExcludeExts: core.ParseExtensionCSV(m.FilterExcludeInput.Value()),
		ExcludeDirs: core.ParseNameCSV(m.FilterDirsInput.Value()),
	}

	selectedRelPaths := m.selectedRelPaths()
	focusedRelPath := m.focusedRelPath()
	filtersChanged := !filterConfigsEqual(m.Filters, newFilters)

	return m.beginAsyncScan(
		"Applying filters...",
		"rescan",
		newFilters,
		nil,
		selectedRelPaths,
		focusedRelPath,
		m.Dirty || filtersChanged,
	)
}

func (m *Model) handleMouse(msg tea.MouseMsg) {
	if m.isCompactLayout() {
		return
	}

	switch msg.Button {
	case tea.MouseButtonWheelUp:
		m.handleMouseWheel(msg.X, msg.Y, -1)
		return

	case tea.MouseButtonWheelDown:
		m.handleMouseWheel(msg.X, msg.Y, 1)
		return
	}

	if msg.Action != tea.MouseActionPress {
		return
	}

	switch msg.Button {
	case tea.MouseButtonLeft:
		m.handleMouseLeftClick(msg.X, msg.Y)
	}
}

func (m *Model) handleMouseWheel(x, y, delta int) {
	pane, ok := m.mousePaneAt(x, y)
	if ok {
		m.ActivePane = pane
	}

	if m.ActivePane == paneInspect {
		if m.scrollInspect(delta) {
			m.StatusMessage = m.inspectStatus()
		}
		return
	}

	if m.moveTreeBy(delta) {
		m.StatusMessage = m.focusStatus()
	}
}

func (m *Model) handleMouseLeftClick(x, y int) {
	navOuter, navContent, inspOuter, _ := m.panelRects()

	if navOuter.Contains(x, y) {
		m.ActivePane = paneTree

		if m.focusTreeRowAt(x, y, navContent.X, navContent.Y, navContent.H) {
			return
		}

		m.StatusMessage = m.focusStatus()
		return
	}

	if inspOuter.Contains(x, y) {
		m.ActivePane = paneInspect
		m.StatusMessage = m.inspectStatus()
	}
}

func (m *Model) focusTreeRowAt(x, y int, contentX, contentY, visibleHeight int) bool {
	if m.Tree == nil || len(m.Tree.VisibleNodes) == 0 {
		return false
	}

	start, end := windowRange(len(m.Tree.VisibleNodes), m.Tree.FocusIndex, visibleHeight)

	row := y - contentY

	if row < 0 {
		return false
	}

	index := start + row
	if index < start || index >= end || index < 0 || index >= len(m.Tree.VisibleNodes) {
		return false
	}

	node := m.Tree.VisibleNodes[index]
	m.Tree.FocusIndex = index

	if m.hitCheckbox(x, contentX, node) {
		if core.ToggleNode(node) {
			m.Dirty = true
			m.refreshPreview()
			m.StatusMessage = m.selectionSummaryStatus()
			return true
		}
	}

	m.refreshPreview()
	m.StatusMessage = m.focusStatus()
	return true
}

func (m Model) hitCheckbox(mouseX, contentX int, node *core.Node) bool {
	if node == nil {
		return false
	}

	checkboxStart := contentX + node.Depth()*2 + 2
	checkboxEnd := checkboxStart + 2

	return mouseX >= checkboxStart && mouseX < checkboxEnd
}

func (m Model) mousePaneAt(x, y int) (paneFocus, bool) {
	navOuter, _, inspOuter, _ := m.panelRects()

	if navOuter.Contains(x, y) {
		return paneTree, true
	}

	if inspOuter.Contains(x, y) {
		return paneInspect, true
	}

	return "", false
}

func (m Model) panelRects() (rect, rect, rect, rect) {
	th := theme.New()
	availableWidth := max(40, m.Width-th.App.GetHorizontalFrameSize())
	gap := 1

	leftWidth := (availableWidth - gap) * 54 / 100
	rightWidth := availableWidth - leftWidth - gap

	bodyY := m.bodyTop()
	bodyH := m.bodyHeight()
	startX := th.App.GetBorderLeftSize() + th.App.GetPaddingLeft()

	panelStyle := th.Panel
	contentOffsetX := panelStyle.GetBorderLeftSize() + panelStyle.GetPaddingLeft()
	contentOffsetY := panelStyle.GetBorderTopSize() + panelStyle.GetPaddingTop()

	navOuter := rect{
		X: startX,
		Y: bodyY,
		W: leftWidth,
		H: bodyH,
	}

	inspOuter := rect{
		X: startX + leftWidth + gap,
		Y: bodyY,
		W: rightWidth,
		H: bodyH,
	}

	navContent := rect{
		X: navOuter.X + contentOffsetX,
		Y: navOuter.Y + contentOffsetY + 2,
		W: panelContentWidth(navOuter.W),
		H: panelContentHeight(navOuter.H),
	}

	inspContent := rect{
		X: inspOuter.X + contentOffsetX,
		Y: inspOuter.Y + contentOffsetY + 2,
		W: panelContentWidth(inspOuter.W),
		H: panelContentHeight(inspOuter.H),
	}

	return navOuter, navContent, inspOuter, inspContent
}

func (m Model) treeVisibleRange() (int, int) {
	if m.Tree == nil || len(m.Tree.VisibleNodes) == 0 {
		return 0, 0
	}

	_, navContent, _, _ := m.panelRects()
	return windowRange(len(m.Tree.VisibleNodes), m.Tree.FocusIndex, navContent.H)
}

func (m *Model) performExport() error {
	if m.Tree == nil || m.Tree.Root == nil {
		return fmt.Errorf("tree not available")
	}

	result, err := core.ExportSelection(m.Tree.Root, core.ExportOptions{
		RootPath:    m.RootPath,
		OutputSpec:  m.OutputSpec,
		ToClipboard: m.ClipboardEnabled,
	}, platform.WriteClipboard)
	if err != nil {
		return err
	}

	m.ResolvedOutputPath = result.OutputPath
	m.Dirty = false
	m.StatusMessage = m.exportStatus(result)
	return nil
}

func (m *Model) moveTreeBy(delta int) bool {
	if m.Tree == nil || delta == 0 {
		return false
	}

	changed := false

	if delta > 0 {
		for i := 0; i < delta; i++ {
			if !m.Tree.MoveDown() {
				break
			}
			changed = true
		}
	} else {
		for i := 0; i < -delta; i++ {
			if !m.Tree.MoveUp() {
				break
			}
			changed = true
		}
	}

	if changed {
		m.refreshPreview()
	}

	return changed
}

func (m *Model) refreshPreview() {
	if m.Tree == nil {
		m.Preview = core.PreviewData{}
		m.InspectScrollOffset = 0
		return
	}

	m.Preview = core.BuildPreview(m.Tree.FocusedNode())
	m.InspectScrollOffset = 0
}

func (m *Model) switchActivePane() {
	if m.ActivePane == paneTree {
		m.ActivePane = paneInspect
		m.clampInspectScroll()
		m.StatusMessage = m.inspectStatus()
		return
	}

	m.ActivePane = paneTree
	m.StatusMessage = m.focusStatus()
}

func (m *Model) scrollInspect(delta int) bool {
	maxOffset := m.inspectMaxScroll()
	newOffset := clamp(m.InspectScrollOffset+delta, 0, maxOffset)

	if newOffset == m.InspectScrollOffset {
		return false
	}

	m.InspectScrollOffset = newOffset
	return true
}

func (m *Model) clampInspectScroll() {
	m.InspectScrollOffset = clamp(m.InspectScrollOffset, 0, m.inspectMaxScroll())
}

func (m Model) inspectMaxScroll() int {
	total := len(m.buildInspectLineItems())
	height := m.inspectViewportHeight()
	if total <= height {
		return 0
	}
	return total - height
}

func (m Model) inspectViewportHeight() int {
	return panelContentHeight(m.bodyHeight())
}

func (m Model) inspectVisibleRange() (int, int, int) {
	total := len(m.buildInspectLineItems())
	if total == 0 {
		return 0, 0, 0
	}

	height := m.inspectViewportHeight()
	offset := clamp(m.InspectScrollOffset, 0, max(0, total-height))
	end := min(total, offset+height)

	return offset, end, total
}

func (m Model) focusStatus() string {
	if m.Tree == nil {
		return "Tree not available."
	}

	node := m.Tree.FocusedNode()
	if node == nil {
		return "No node focused."
	}

	kind := "file"
	if node.IsDir {
		kind = "folder"
	}

	return fmt.Sprintf(
		"Navigator · focused %s: %s · state: %s · selected files: %d",
		kind,
		node.RelPath,
		core.SelectionState(node),
		m.selectedFilesCount(),
	)
}

func (m Model) inspectStatus() string {
	start, end, total := m.inspectVisibleRange()

	node := "<none>"
	if m.Tree != nil && m.Tree.FocusedNode() != nil {
		node = m.Tree.FocusedNode().RelPath
	}

	if total == 0 {
		return fmt.Sprintf("Inspect · %s · no content", node)
	}

	return fmt.Sprintf(
		"Inspect · %s · lines %d-%d/%d",
		node,
		start+1,
		end,
		total,
	)
}

func (m Model) selectionSummaryStatus() string {
	if m.Tree == nil || m.Tree.Root == nil {
		return "No tree available."
	}

	node := m.Tree.FocusedNode()
	if node == nil {
		return fmt.Sprintf("Selected %d file(s).", m.selectedFilesCount())
	}

	return fmt.Sprintf(
		"Toggled %s · state: %s · selected files: %d",
		node.RelPath,
		core.SelectionState(node),
		m.selectedFilesCount(),
	)
}

func (m Model) exportStatus(result core.ExportResult) string {
	status := fmt.Sprintf(
		"Exported %d/%d file(s) to tools/output.txt",
		result.ExportedCount,
		result.SelectedCount,
	)

	if result.SkippedCount > 0 {
		status += fmt.Sprintf(" · skipped %d unreadable", result.SkippedCount)
	}

	if result.Clipboard {
		status += " · clipboard updated"
	}

	if result.GitIgnorePatched {
		status += " · .gitignore patched"
	}

	return status
}

func (m Model) selectedFilesCount() int {
	if m.Tree == nil || m.Tree.Root == nil {
		return 0
	}

	return core.CountSelectedFiles(m.Tree.Root)
}

func (m Model) selectedRelPaths() []string {
	if m.Tree == nil || m.Tree.Root == nil {
		return nil
	}

	selectedNodes := core.CollectSelectedFiles(m.Tree.Root)
	paths := make([]string, 0, len(selectedNodes))
	for _, node := range selectedNodes {
		paths = append(paths, node.RelPath)
	}

	return paths
}

func (m Model) focusedRelPath() string {
	if m.Tree == nil || m.Tree.FocusedNode() == nil {
		return ""
	}
	return m.Tree.FocusedNode().RelPath
}

func (m Model) filterBadgeLabel() string {
	if m.hasCustomFilters() {
		return "[f]ilters: custom"
	}
	return "[f]ilters: base"
}

func (m Model) hasCustomFilters() bool {
	return !filterConfigsEqual(m.Filters, core.DefaultFilterConfig())
}

func (m Model) layoutMetrics() (headerHeight, bodyHeight, footerHeight int) {
	th := theme.New()

	headerHeight = lipgloss.Height(m.renderHeader(th))
	footerHeight = lipgloss.Height(m.renderFooter(th))

	bodyHeight = m.Height - th.App.GetVerticalFrameSize() - headerHeight - footerHeight
	if bodyHeight < 10 {
		bodyHeight = 10
	}

	return headerHeight, bodyHeight, footerHeight
}

func (m Model) bodyHeight() int {
	_, bodyHeight, _ := m.layoutMetrics()
	return bodyHeight
}

func (m Model) bodyTop() int {
	th := theme.New()
	headerHeight, _, _ := m.layoutMetrics()
	return th.App.GetBorderTopSize() + th.App.GetPaddingTop() + headerHeight
}

func (m *Model) initFilterInputs() {
	m.FilterIncludeInput = newFilterTextInput(
		core.FormatSetCSV(m.Filters.IncludeExts),
		".go,.py,.md",
	)
	m.FilterExcludeInput = newFilterTextInput(
		core.FormatSetCSV(m.Filters.ExcludeExts),
		".png,.jpg,.exe",
	)
	m.FilterDirsInput = newFilterTextInput(
		core.FormatSetCSV(m.Filters.ExcludeDirs),
		"node_modules,dist,build",
	)

	m.FilterFocusIndex = 0
	m.blurAllFilterInputs()
}

func (m *Model) nativePickCmd(target string) tea.Cmd {
	startDir := ""
	if target == "root" {
		startDir = m.RootPath
	} else {
		// output: start in directory of resolved output
		if m.ResolvedOutputPath != "" {
			startDir = filepath.Dir(m.ResolvedOutputPath)
		} else {
			startDir = m.RootPath
		}
	}

	return func() tea.Msg {
		if !platform.SupportsNativePicker() {
			return nativePickResultMsg{Target: target, Cancelled: true}
		}

		if target == "root" {
			p, cancelled, err := platform.PickFolder(startDir)
			return nativePickResultMsg{Target: target, Path: p, Cancelled: cancelled, Err: err}
		}

		p, cancelled, err := platform.PickSaveFile(startDir, "output.txt")
		return nativePickResultMsg{Target: target, Path: p, Cancelled: cancelled, Err: err}
	}
}

func newFilterTextInput(value, placeholder string) textinput.Model {
	input := textinput.New()
	input.Prompt = ""
	input.Placeholder = placeholder
	input.SetValue(value)
	input.CharLimit = 512
	input.Width = 48
	return input
}

func (m *Model) openFilters() {
	m.syncFilterInputsFromConfig()
	m.ShowFilters = true
	m.FilterFocusIndex = 0
	m.focusFilterInput(0)
	m.StatusMessage = "Filters drawer opened."
}

func (m *Model) closeFilters(status string) {
	m.ShowFilters = false
	m.blurAllFilterInputs()
	m.StatusMessage = status
}

func (m *Model) syncFilterInputsFromConfig() {
	m.FilterIncludeInput.SetValue(core.FormatSetCSV(m.Filters.IncludeExts))
	m.FilterExcludeInput.SetValue(core.FormatSetCSV(m.Filters.ExcludeExts))
	m.FilterDirsInput.SetValue(core.FormatSetCSV(m.Filters.ExcludeDirs))
}

func (m *Model) resetFilterInputsToDefault() {
	defaults := core.DefaultFilterConfig()
	m.FilterIncludeInput.SetValue(core.FormatSetCSV(defaults.IncludeExts))
	m.FilterExcludeInput.SetValue(core.FormatSetCSV(defaults.ExcludeExts))
	m.FilterDirsInput.SetValue(core.FormatSetCSV(defaults.ExcludeDirs))
}

func (m *Model) cycleFilterFocus(delta int) {
	total := 3
	m.FilterFocusIndex = (m.FilterFocusIndex + delta + total) % total
	m.focusFilterInput(m.FilterFocusIndex)
}

func (m *Model) focusFilterInput(index int) {
	m.blurAllFilterInputs()

	switch index {
	case 0:
		m.FilterIncludeInput.Focus()
	case 1:
		m.FilterExcludeInput.Focus()
	case 2:
		m.FilterDirsInput.Focus()
	}
}

func (m *Model) blurAllFilterInputs() {
	m.FilterIncludeInput.Blur()
	m.FilterExcludeInput.Blur()
	m.FilterDirsInput.Blur()
}

func (m *Model) updateFocusedFilterInput(msg tea.Msg) tea.Cmd {
	switch m.FilterFocusIndex {
	case 0:
		var cmd tea.Cmd
		m.FilterIncludeInput, cmd = m.FilterIncludeInput.Update(msg)
		return cmd
	case 1:
		var cmd tea.Cmd
		m.FilterExcludeInput, cmd = m.FilterExcludeInput.Update(msg)
		return cmd
	case 2:
		var cmd tea.Cmd
		m.FilterDirsInput, cmd = m.FilterDirsInput.Update(msg)
		return cmd
	default:
		return nil
	}
}

func filterConfigsEqual(a, b core.FilterConfig) bool {
	return stringSetEqual(a.IncludeExts, b.IncludeExts) &&
		stringSetEqual(a.ExcludeExts, b.ExcludeExts) &&
		stringSetEqual(a.ExcludeDirs, b.ExcludeDirs)
}

func stringSetEqual(a, b map[string]bool) bool {
	if len(a) != len(b) {
		return false
	}

	for key := range a {
		if !b[key] {
			return false
		}
	}

	return true
}

func clamp(value, low, high int) int {
	if value < low {
		return low
	}
	if value > high {
		return high
	}
	return value
}

func centeredOffset(container, content int) int {
	gap := container - content
	if gap <= 0 {
		return 0
	}

	// Match lipgloss center behavior for placement offsets.
	return gap / 2
}

func (m Model) drawerPanelAndContentRects(style lipgloss.Style, panelW, panelH int) (rect, int, int) {
	placeW := max(40, m.Width-2)
	placeH := max(12, m.Height-2)

	renderedW := panelW + style.GetHorizontalFrameSize()
	renderedH := panelH + style.GetVerticalFrameSize()

	left := centeredOffset(placeW, renderedW)
	top := centeredOffset(placeH, renderedH)

	panelRect := rect{
		X: left,
		Y: top,
		W: renderedW,
		H: renderedH,
	}

	contentX := panelRect.X + style.GetBorderLeftSize() + style.GetPaddingLeft()
	contentY := panelRect.Y + style.GetBorderTopSize() + style.GetPaddingTop()

	return panelRect, contentX, contentY
}

func (m *Model) initWorkspaceInputs() {
	m.WorkspaceRootInput = newWorkspaceTextInput(m.RootPath, `D:\codes\myproject`)
	m.WorkspaceOutputInput = newWorkspaceTextInput(m.OutputSpec, `. or C:\temp\merged.md`)

	m.WorkspaceFocusIndex = 0
	m.blurWorkspaceInputs()
}

func (m *Model) initPalette() {
	in := textinput.New()
	in.Prompt = ""
	in.Placeholder = "search paths (e.g. src/main, README, docs api)"
	in.CharLimit = 256
	in.Width = 60
	m.PaletteInput = in
}

func newWorkspaceTextInput(value, placeholder string) textinput.Model {
	input := textinput.New()
	input.Prompt = ""
	input.Placeholder = placeholder
	input.SetValue(value)
	input.CharLimit = 1024
	input.Width = 64
	return input
}

func (m *Model) openWorkspace() {
	m.syncWorkspaceInputsFromState()
	m.ShowWorkspace = true
	m.WorkspaceFocusIndex = 0
	m.focusWorkspaceInput(0)
	m.StatusMessage = "Workspace drawer opened."
}

func (m *Model) closeWorkspace(status string) {
	m.ShowWorkspace = false
	m.blurWorkspaceInputs()
	m.StatusMessage = status
}

func (m *Model) syncWorkspaceInputsFromState() {
	m.WorkspaceRootInput.SetValue(m.RootPath)
	m.WorkspaceOutputInput.SetValue(m.OutputSpec)
}

func (m *Model) resetWorkspaceInputsToCurrent() {
	m.syncWorkspaceInputsFromState()
}

func (m *Model) cycleWorkspaceFocus(delta int) {
	total := 2
	m.WorkspaceFocusIndex = (m.WorkspaceFocusIndex + delta + total) % total
	m.focusWorkspaceInput(m.WorkspaceFocusIndex)
}

func (m *Model) focusWorkspaceInput(index int) {
	m.blurWorkspaceInputs()

	switch index {
	case 0:
		m.WorkspaceRootInput.Focus()
	case 1:
		m.WorkspaceOutputInput.Focus()
	}
}

func (m *Model) blurWorkspaceInputs() {
	m.WorkspaceRootInput.Blur()
	m.WorkspaceOutputInput.Blur()
}

func (m *Model) updateFocusedWorkspaceInput(msg tea.Msg) tea.Cmd {
	switch m.WorkspaceFocusIndex {
	case 0:
		var cmd tea.Cmd
		m.WorkspaceRootInput, cmd = m.WorkspaceRootInput.Update(msg)
		return cmd
	case 1:
		var cmd tea.Cmd
		m.WorkspaceOutputInput, cmd = m.WorkspaceOutputInput.Update(msg)
		return cmd
	default:
		return nil
	}
}

func (m Model) workspaceResolvedOutputPreview() (string, error) {
	rootValue := strings.TrimSpace(m.WorkspaceRootInput.Value())
	if rootValue == "" {
		return "", fmt.Errorf("root path cannot be empty")
	}

	absRoot, err := filepath.Abs(rootValue)
	if err != nil {
		return "", fmt.Errorf("resolve root path: %w", err)
	}

	return core.ResolveOutputPath(absRoot, strings.TrimSpace(m.WorkspaceOutputInput.Value()))
}

func (m *Model) applyWorkspace() (tea.Cmd, error) {
	rootValue := strings.TrimSpace(m.WorkspaceRootInput.Value())
	if rootValue == "" {
		return nil, fmt.Errorf("root path cannot be empty")
	}

	absRoot, err := filepath.Abs(rootValue)
	if err != nil {
		return nil, fmt.Errorf("resolve root path: %w", err)
	}

	info, err := os.Stat(absRoot)
	if err != nil {
		return nil, fmt.Errorf("stat root path: %w", err)
	}
	if !info.IsDir() {
		return nil, fmt.Errorf("root must be a directory: %s", absRoot)
	}

	outputSpec := strings.TrimSpace(m.WorkspaceOutputInput.Value())
	resolvedOutputPath, err := core.ResolveOutputPath(absRoot, outputSpec)
	if err != nil {
		return nil, fmt.Errorf("resolve output target: %w", err)
	}

	rootChanged := !samePath(m.RootPath, absRoot)
	outputChanged := !samePath(m.ResolvedOutputPath, resolvedOutputPath)

	if !rootChanged && !outputChanged {
		m.closeWorkspace("Workspace unchanged.")
		return nil, nil
	}

	if !rootChanged {
		m.OutputSpec = outputSpec
		m.ResolvedOutputPath = resolvedOutputPath
		m.Dirty = true
		m.closeWorkspace("Workspace updated.")
		return nil, nil
	}

	m.RootPath = absRoot
	m.OutputSpec = outputSpec
	m.ResolvedOutputPath = resolvedOutputPath
	m.InitialSelections = nil
	m.ActivePane = paneTree
	m.closeWorkspace("Switching workspace...")

	cmd := m.beginAsyncScan(
		"Switching workspace...",
		"workspace",
		m.Filters,
		nil,
		nil,
		"",
		false,
	)

	return cmd, nil
}

func samePath(left, right string) bool {
	a := filepath.Clean(left)
	b := filepath.Clean(right)

	if runtime.GOOS == "windows" {
		return strings.EqualFold(a, b)
	}

	return a == b
}
