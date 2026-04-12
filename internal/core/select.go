package core

import (
	"path/filepath"
	"runtime"
	"strings"
)

func ToggleNode(node *Node) bool {
	if node == nil {
		return false
	}

	if node.IsDir {
		target := true
		if node.Selected && !node.Partial {
			target = false
		}

		SetSubtreeSelected(node, target)
		RecalculateAncestors(node)
		return true
	}

	node.Selected = !node.Selected
	node.Partial = false
	RecalculateAncestors(node)
	return true
}

func SelectNode(node *Node) {
	if node == nil {
		return
	}

	SetSubtreeSelected(node, true)
	RecalculateAncestors(node)
}

func SetSubtreeSelected(node *Node, selected bool) {
	if node == nil {
		return
	}

	node.Selected = selected
	node.Partial = false

	for _, child := range node.Children {
		SetSubtreeSelected(child, selected)
	}
}

func RecalculateAncestors(node *Node) {
	for current := node.Parent; current != nil; current = current.Parent {
		recalculateNodeState(current)
	}
}

func recalculateNodeState(node *Node) {
	if node == nil || !node.IsDir {
		return
	}

	if len(node.Children) == 0 {
		node.Partial = false
		return
	}

	allSelected := true
	anySelected := false

	for _, child := range node.Children {
		if child.Partial {
			anySelected = true
			allSelected = false
			continue
		}

		if child.Selected {
			anySelected = true
		} else {
			allSelected = false
		}
	}

	switch {
	case allSelected && anySelected:
		node.Selected = true
		node.Partial = false
	case anySelected:
		node.Selected = false
		node.Partial = true
	default:
		node.Selected = false
		node.Partial = false
	}
}

func CountSelectedFiles(node *Node) int {
	if node == nil {
		return 0
	}

	if !node.IsDir {
		if node.Selected {
			return 1
		}
		return 0
	}

	total := 0
	for _, child := range node.Children {
		total += CountSelectedFiles(child)
	}

	return total
}

func CountSelectedNodes(node *Node) int {
	if node == nil {
		return 0
	}

	total := 0
	if node.Selected || node.Partial {
		total++
	}

	for _, child := range node.Children {
		total += CountSelectedNodes(child)
	}

	return total
}

func SelectionState(node *Node) string {
	if node == nil {
		return "unselected"
	}

	if node.Partial {
		return "partial"
	}

	if node.Selected {
		return "selected"
	}

	return "unselected"
}

func ApplyInitialSelections(root *Node, rootPath string, rawPaths []string) (int, []string) {
	if root == nil || len(rawPaths) == 0 {
		return 0, nil
	}

	index := indexNodesByAbsPath(root)
	applied := 0
	var skipped []string

	for _, raw := range rawPaths {
		trimmed := strings.TrimSpace(raw)
		if trimmed == "" {
			continue
		}

		absTarget := trimmed
		if !filepath.IsAbs(absTarget) {
			absTarget = filepath.Join(rootPath, absTarget)
		}

		absTarget, err := filepath.Abs(absTarget)
		if err != nil {
			skipped = append(skipped, raw)
			continue
		}

		key := normalizePathKey(absTarget)
		node, ok := index[key]
		if !ok {
			skipped = append(skipped, raw)
			continue
		}

		SelectNode(node)
		applied++
	}

	return applied, skipped
}

func indexNodesByAbsPath(root *Node) map[string]*Node {
	index := map[string]*Node{}

	var walk func(*Node)
	walk = func(node *Node) {
		if node == nil {
			return
		}

		index[normalizePathKey(node.AbsPath)] = node

		for _, child := range node.Children {
			walk(child)
		}
	}

	walk(root)
	return index
}

func normalizePathKey(path string) string {
	clean := filepath.Clean(path)

	if runtime.GOOS == "windows" {
		return strings.ToLower(clean)
	}

	return clean
}