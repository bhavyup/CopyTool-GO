package core

import (
	"os"
	"path/filepath"
	"testing"
)

func TestToggleNodeDirectorySelectsAndUnselectsSubtree(t *testing.T) {
	t.Parallel()

	root, src, mainGo, readme := buildSelectionFixture(t)

	if !ToggleNode(src) {
		t.Fatalf("expected toggle to succeed")
	}

	if !src.Selected || src.Partial {
		t.Fatalf("expected src directory to become fully selected")
	}
	if !mainGo.Selected {
		t.Fatalf("expected file under src to be selected")
	}
	if readme.Selected {
		t.Fatalf("expected unrelated sibling file to stay unselected")
	}
	if !root.Partial || root.Selected {
		t.Fatalf("expected root to become partial when only part of subtree is selected")
	}

	if !ToggleNode(src) {
		t.Fatalf("expected second toggle to succeed")
	}

	if src.Selected || src.Partial || mainGo.Selected || root.Partial || root.Selected {
		t.Fatalf("expected subtree and ancestors to return to unselected state")
	}
}

func TestApplyInitialSelections(t *testing.T) {
	t.Parallel()

	root, _, mainGo, readme := buildSelectionFixture(t)
	rootPath := root.AbsPath

	applied, skipped := ApplyInitialSelections(root, rootPath, []string{
		"src/main.go",
		"README.md",
		"missing.txt",
	})

	if applied != 2 {
		t.Fatalf("unexpected applied count: got %d want %d", applied, 2)
	}
	if len(skipped) != 1 || skipped[0] != "missing.txt" {
		t.Fatalf("unexpected skipped paths: %v", skipped)
	}

	if !mainGo.Selected || !readme.Selected {
		t.Fatalf("expected targeted files to be selected")
	}
}

func buildSelectionFixture(t *testing.T) (root, src, mainGo, readme *Node) {
	t.Helper()

	base := t.TempDir()
	if err := os.MkdirAll(filepath.Join(base, "src"), 0o755); err != nil {
		t.Fatalf("create src directory: %v", err)
	}
	if err := os.WriteFile(filepath.Join(base, "src", "main.go"), []byte("package main\n"), 0o644); err != nil {
		t.Fatalf("create src/main.go: %v", err)
	}
	if err := os.WriteFile(filepath.Join(base, "README.md"), []byte("# CopyTool\n"), 0o644); err != nil {
		t.Fatalf("create README.md: %v", err)
	}

	root = &Node{
		Name:    filepath.Base(base),
		RelPath: ".",
		AbsPath: base,
		IsDir:   true,
	}

	src = &Node{
		Name:    "src",
		RelPath: "src",
		AbsPath: filepath.Join(base, "src"),
		IsDir:   true,
		Parent:  root,
	}

	mainGo = &Node{
		Name:    "main.go",
		RelPath: filepath.Join("src", "main.go"),
		AbsPath: filepath.Join(base, "src", "main.go"),
		IsDir:   false,
		Ext:     ".go",
		Parent:  src,
	}

	readme = &Node{
		Name:    "README.md",
		RelPath: "README.md",
		AbsPath: filepath.Join(base, "README.md"),
		IsDir:   false,
		Ext:     ".md",
		Parent:  root,
	}

	src.Children = []*Node{mainGo}
	root.Children = []*Node{src, readme}

	return root, src, mainGo, readme
}
