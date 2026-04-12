package core

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

var errSkipPath = errors.New("skip path")

func ScanRoot(root string, filters FilterConfig) (*Tree, error) {
	absRoot, err := filepath.Abs(root)
	if err != nil {
		return nil, fmt.Errorf("resolve root path: %w", err)
	}

	info, err := os.Stat(absRoot)
	if err != nil {
		return nil, fmt.Errorf("stat root path %s: %w", absRoot, err)
	}

	if !info.IsDir() {
		return nil, fmt.Errorf("root must be a directory: %s", absRoot)
	}

	rootNode, err := buildNode(absRoot, ".", nil, true, filters, newIgnoreState(nil))
	if err != nil {
		return nil, err
	}

	rootNode.Expanded = true
	return NewTree(rootNode), nil
}

func buildNode(absPath, relPath string, parent *Node, isRoot bool, filters FilterConfig, state ignoreState) (*Node, error) {
	info, err := os.Lstat(absPath)
	if err != nil {
		if !isRoot && isSkippableScanError(err) {
			return nil, skipPathError(absPath, err)
		}
		return nil, fmt.Errorf("read path info %s: %w", absPath, err)
	}

	if info.Mode()&os.ModeSymlink != 0 {
		return nil, skipPathError(absPath, fmt.Errorf("symlink/reparse point skipped"))
	}

	node := &Node{
		Name:    filepath.Base(absPath),
		RelPath: relPath,
		AbsPath: absPath,
		IsDir:   info.IsDir(),
		Parent:  parent,
		Size:    info.Size(),
		Ext:     strings.ToLower(filepath.Ext(absPath)),
	}

	if isRoot {
		node.Name = filepath.Base(filepath.Clean(absPath))
	}

	if !node.IsDir {
		return node, nil
	}

	activeState, err := state.ExtendForDirectory(absPath, relPath)
	if err != nil {
		if !isRoot && isSkippableScanError(err) {
			return nil, skipPathError(absPath, err)
		}
		return nil, fmt.Errorf("load .gitignore in %s: %w", absPath, err)
	}

	entries, err := os.ReadDir(absPath)
	if err != nil {
		if !isRoot && isSkippableScanError(err) {
			return nil, skipPathError(absPath, err)
		}
		return nil, fmt.Errorf("read directory %s: %w", absPath, err)
	}

	sort.Slice(entries, func(i, j int) bool {
		leftDir := entries[i].IsDir()
		rightDir := entries[j].IsDir()

		if leftDir != rightDir {
			return leftDir
		}

		return strings.ToLower(entries[i].Name()) < strings.ToLower(entries[j].Name())
	})

	for _, entry := range entries {
		name := entry.Name()
		childAbs := filepath.Join(absPath, name)

		childInfo, err := os.Lstat(childAbs)
		if err != nil {
			if isSkippableScanError(err) {
				continue
			}
			return nil, fmt.Errorf("read path info %s: %w", childAbs, err)
		}

		if childInfo.Mode()&os.ModeSymlink != 0 {
			continue
		}

		isDir := childInfo.IsDir()
		lowerName := strings.ToLower(name)

		if isDir {
			if IsHardExcludedDir(lowerName) || filters.ExcludeDirs[lowerName] {
				continue
			}
		} else {
			if !FilePassesExtensionFilters(name, filters) {
				continue
			}
		}

		childRel := name
		if relPath != "." {
			childRel = filepath.Join(relPath, name)
		}

		if activeState.Match(childRel, isDir) {
			continue
		}

		childNode, err := buildNode(childAbs, childRel, node, false, filters, activeState)
		if err != nil {
			if errors.Is(err, errSkipPath) {
				continue
			}
			return nil, err
		}

		if childNode != nil {
			node.Children = append(node.Children, childNode)
		}
	}

	return node, nil
}

func skipPathError(path string, err error) error {
	return fmt.Errorf("%w: %s: %v", errSkipPath, path, err)
}

func isSkippableScanError(err error) bool {
	if err == nil {
		return false
	}

	if errors.Is(err, fs.ErrPermission) || errors.Is(err, fs.ErrNotExist) {
		return true
	}

	msg := strings.ToLower(err.Error())

	switch {
	case strings.Contains(msg, "access is denied"):
		return true
	case strings.Contains(msg, "permission denied"):
		return true
	case strings.Contains(msg, "cannot access the file"):
		return true
	case strings.Contains(msg, "the process cannot access the file"):
		return true
	case strings.Contains(msg, "path not found"):
		return true
	default:
		return false
	}
}