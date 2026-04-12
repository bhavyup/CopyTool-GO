package tui

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

type pickerEntry struct {
	Label string
	Path  string
	IsDir bool
	Kind  string // current, parent, dir, file
}

func (m *Model) openPickerForWorkspaceField() error {
	mode := "root"
	if m.WorkspaceFocusIndex == 1 {
		mode = "output"
	}

	startPath, err := m.resolvePickerStartPath(mode)
	if err != nil {
		return err
	}

	entries, err := readPickerEntries(startPath, mode)
	if err != nil {
		return err
	}

	m.ShowPicker = true
	m.PickerMode = mode
	m.PickerPath = startPath
	m.PickerEntries = entries
	m.PickerIndex = 0
	m.StatusMessage = "Picker opened."

	return nil
}

func (m *Model) closePicker(status string) {
	m.ShowPicker = false
	m.PickerMode = ""
	m.PickerPath = ""
	m.PickerEntries = nil
	m.PickerIndex = 0
	m.StatusMessage = status
}

func (m *Model) resolvePickerStartPath(mode string) (string, error) {
	raw := ""
	switch mode {
	case "root":
		raw = strings.TrimSpace(m.WorkspaceRootInput.Value())
	case "output":
		raw = strings.TrimSpace(m.WorkspaceOutputInput.Value())
	}

	if raw == "" {
		if mode == "root" {
			return m.RootPath, nil
		}
		return m.RootPath, nil
	}

	if raw == "." && mode == "output" {
		return os.Getwd()
	}

	abs, err := filepath.Abs(raw)
	if err != nil {
		return "", fmt.Errorf("resolve picker path: %w", err)
	}

	info, err := os.Stat(abs)
	if err == nil {
		if info.IsDir() {
			return abs, nil
		}
		return filepath.Dir(abs), nil
	}

	if !os.IsNotExist(err) {
		return "", err
	}

	if ext := filepath.Ext(filepath.Base(abs)); ext != "" && mode == "output" {
		return filepath.Dir(abs), nil
	}

	return abs, nil
}

func readPickerEntries(dir string, mode string) ([]pickerEntry, error) {
	entries := []pickerEntry{
		{
			Label: "[.] choose current directory",
			Path:  dir,
			IsDir: true,
			Kind:  "current",
		},
	}

	parent := filepath.Dir(dir)
	if parent != dir {
		entries = append(entries, pickerEntry{
			Label: "[..] parent directory",
			Path:  parent,
			IsDir: true,
			Kind:  "parent",
		})
	}

	dirEntries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	var dirs []pickerEntry
	var files []pickerEntry

	for _, entry := range dirEntries {
		name := entry.Name()
		full := filepath.Join(dir, name)

		info, err := os.Lstat(full)
		if err != nil {
			continue
		}

		if info.Mode()&os.ModeSymlink != 0 {
			continue
		}

		if info.IsDir() {
			dirs = append(dirs, pickerEntry{
				Label: name + "/",
				Path:  full,
				IsDir: true,
				Kind:  "dir",
			})
			continue
		}

		if mode == "output" {
			files = append(files, pickerEntry{
				Label: name,
				Path:  full,
				IsDir: false,
				Kind:  "file",
			})
		}
	}

	sort.Slice(dirs, func(i, j int) bool {
		return strings.ToLower(dirs[i].Label) < strings.ToLower(dirs[j].Label)
	})

	sort.Slice(files, func(i, j int) bool {
		return strings.ToLower(files[i].Label) < strings.ToLower(files[j].Label)
	})

	entries = append(entries, dirs...)
	entries = append(entries, files...)

	return entries, nil
}

func (m *Model) pickerMove(delta int) {
	if len(m.PickerEntries) == 0 {
		return
	}

	m.PickerIndex += delta
	if m.PickerIndex < 0 {
		m.PickerIndex = 0
	}
	if m.PickerIndex >= len(m.PickerEntries) {
		m.PickerIndex = len(m.PickerEntries) - 1
	}
}

func (m Model) pickerCurrentEntry() *pickerEntry {
	if len(m.PickerEntries) == 0 {
		return nil
	}
	if m.PickerIndex < 0 || m.PickerIndex >= len(m.PickerEntries) {
		return nil
	}
	return &m.PickerEntries[m.PickerIndex]
}

func (m *Model) pickerOpenOrSelect() error {
	entry := m.pickerCurrentEntry()
	if entry == nil {
		return nil
	}

	switch entry.Kind {
	case "current":
		return m.applyPickedPath(entry.Path)

	case "parent":
		return m.loadPickerDirectory(entry.Path)

	case "dir":
		return m.loadPickerDirectory(entry.Path)

	case "file":
		if m.PickerMode == "output" {
			return m.applyPickedPath(entry.Path)
		}
	}

	return nil
}

func (m *Model) pickerSelectHighlighted() error {
	entry := m.pickerCurrentEntry()
	if entry == nil {
		return nil
	}

	switch entry.Kind {
	case "current", "parent", "dir":
		return m.applyPickedPath(entry.Path)
	case "file":
		if m.PickerMode == "output" {
			return m.applyPickedPath(entry.Path)
		}
	}

	return nil
}

func (m *Model) pickerGoParent() error {
	if m.PickerPath == "" {
		return nil
	}

	parent := filepath.Dir(m.PickerPath)
	if parent == m.PickerPath {
		return nil
	}

	return m.loadPickerDirectory(parent)
}

func (m *Model) loadPickerDirectory(dir string) error {
	entries, err := readPickerEntries(dir, m.PickerMode)
	if err != nil {
		return err
	}

	m.PickerPath = dir
	m.PickerEntries = entries
	m.PickerIndex = 0
	return nil
}

func (m *Model) applyPickedPath(path string) error {
	switch m.PickerMode {
	case "root":
		info, err := os.Stat(path)
		if err != nil {
			return err
		}
		if !info.IsDir() {
			return fmt.Errorf("root must be a directory")
		}
		m.WorkspaceRootInput.SetValue(path)
		m.closePicker("Picker selected root path.")
		m.focusWorkspaceInput(0)
		return nil

	case "output":
		m.WorkspaceOutputInput.SetValue(path)
		m.closePicker("Picker selected output target.")
		m.focusWorkspaceInput(1)
		return nil
	}

	return nil
}