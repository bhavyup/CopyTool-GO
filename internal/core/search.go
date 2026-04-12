package core

import (
	"path/filepath"
	"runtime"
	"strings"
)

func FindNodeByRelPath(root *Node, relPath string) *Node {
	if root == nil {
		return nil
	}

	target := normalizeRelPath(relPath)

	var found *Node
	var walk func(*Node)
	walk = func(node *Node) {
		if node == nil || found != nil {
			return
		}

		if normalizeRelPath(node.RelPath) == target {
			found = node
			return
		}

		for _, child := range node.Children {
			walk(child)
		}
	}

	walk(root)
	return found
}

func ExpandAncestors(node *Node) {
	for current := node; current != nil; current = current.Parent {
		if current.IsDir {
			current.Expanded = true
		}
	}
}

func normalizeRelPath(relPath string) string {
	clean := filepath.ToSlash(filepath.Clean(relPath))
	if runtime.GOOS == "windows" {
		return strings.ToLower(clean)
	}
	return clean
}