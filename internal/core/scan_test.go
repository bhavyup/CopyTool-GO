package core

import (
	"os"
	"path/filepath"
	"testing"
)

func TestScanRootAppliesGitignoreAndDirectoryFilters(t *testing.T) {
	t.Parallel()

	root := t.TempDir()

	mustWriteFile(t, filepath.Join(root, ".gitignore"), "README.md\nignored_dir/\nsrc/*.tmp\n")
	mustWriteFile(t, filepath.Join(root, "main.go"), "package main\n")
	mustWriteFile(t, filepath.Join(root, "README.md"), "# docs\n")
	mustWriteFile(t, filepath.Join(root, "src", "app.go"), "package src\n")
	mustWriteFile(t, filepath.Join(root, "src", "app.tmp"), "tmp\n")
	mustWriteFile(t, filepath.Join(root, "src", ".gitignore"), "local-ignore.go\n")
	mustWriteFile(t, filepath.Join(root, "src", "local-ignore.go"), "package src\n")
	mustWriteFile(t, filepath.Join(root, "src", "keep.go"), "package src\n")
	mustWriteFile(t, filepath.Join(root, "ignored_dir", "hidden.go"), "package ignored\n")
	mustWriteFile(t, filepath.Join(root, ".git", "config"), "[core]\n")
	mustWriteFile(t, filepath.Join(root, "tools", "list.txt"), "src/app.go\n")
	mustWriteFile(t, filepath.Join(root, "dist", "generated.go"), "package dist\n")

	tree, err := ScanRoot(root, DefaultFilterConfig())
	if err != nil {
		t.Fatalf("ScanRoot returned error: %v", err)
	}

	paths := collectRelPaths(tree.Root)

	for _, expected := range []string{".", "main.go", "src", "src/app.go", "src/keep.go"} {
		if !paths[expected] {
			t.Fatalf("expected path %q to be visible", expected)
		}
	}

	for _, blocked := range []string{"README.md", "ignored_dir", "ignored_dir/hidden.go", "src/app.tmp", "src/local-ignore.go", ".git", "tools", "dist"} {
		if paths[blocked] {
			t.Fatalf("expected path %q to be excluded", blocked)
		}
	}
}

func TestScanRootRespectsIncludeExtensions(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	mustWriteFile(t, filepath.Join(root, "keep.go"), "package main\n")
	mustWriteFile(t, filepath.Join(root, "drop.md"), "# docs\n")

	filters := DefaultFilterConfig()
	filters.IncludeExts = ParseExtensionCSV("go")

	tree, err := ScanRoot(root, filters)
	if err != nil {
		t.Fatalf("ScanRoot returned error: %v", err)
	}

	paths := collectRelPaths(tree.Root)
	if !paths["keep.go"] {
		t.Fatalf("expected keep.go to pass include filter")
	}
	if paths["drop.md"] {
		t.Fatalf("expected drop.md to be filtered out")
	}
}

func collectRelPaths(root *Node) map[string]bool {
	out := map[string]bool{}

	var walk func(*Node)
	walk = func(node *Node) {
		if node == nil {
			return
		}
		out[filepath.ToSlash(node.RelPath)] = true
		for _, child := range node.Children {
			walk(child)
		}
	}

	walk(root)
	return out
}

func mustWriteFile(t *testing.T, path string, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("create parent directory for %s: %v", path, err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write file %s: %v", path, err)
	}
}
