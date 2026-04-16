package core

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestExportSelectionWritesDeterministicArtifacts(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	mustWriteFile(t, filepath.Join(root, "a.go"), "package main\n")
	mustWriteFile(t, filepath.Join(root, "b.md"), "# Title\n")
	mustWriteFile(t, filepath.Join(root, ".gitignore"), "node_modules/\n")

	rootNode := &Node{
		Name:    filepath.Base(root),
		RelPath: ".",
		AbsPath: root,
		IsDir:   true,
	}

	nodeB := &Node{
		Name:     "b.md",
		RelPath:  "b.md",
		AbsPath:  filepath.Join(root, "b.md"),
		Ext:      ".md",
		Selected: true,
		Parent:   rootNode,
	}
	nodeA := &Node{
		Name:     "a.go",
		RelPath:  "a.go",
		AbsPath:  filepath.Join(root, "a.go"),
		Ext:      ".go",
		Selected: true,
		Parent:   rootNode,
	}
	rootNode.Children = []*Node{nodeB, nodeA}

	var clipboard string
	result, err := ExportSelection(rootNode, ExportOptions{
		RootPath:    root,
		ToClipboard: true,
	}, func(data string) error {
		clipboard = data
		return nil
	})
	if err != nil {
		t.Fatalf("ExportSelection returned error: %v", err)
	}

	if result.SelectedCount != 2 || result.ExportedCount != 2 || result.SkippedCount != 0 {
		t.Fatalf("unexpected result counts: %+v", result)
	}
	if !result.GitIgnorePatched {
		t.Fatalf("expected .gitignore to be patched with tools/")
	}

	listBytes, err := os.ReadFile(result.ListPath)
	if err != nil {
		t.Fatalf("read generated list file: %v", err)
	}
	if string(listBytes) != "a.go\nb.md\n" {
		t.Fatalf("unexpected list file content: %q", string(listBytes))
	}

	outputBytes, err := os.ReadFile(result.OutputPath)
	if err != nil {
		t.Fatalf("read generated output file: %v", err)
	}
	output := string(outputBytes)

	aBlock := "a.go\n````go\n"
	bBlock := "b.md\n````markdown\n"
	aIdx := strings.Index(output, aBlock)
	bIdx := strings.Index(output, bBlock)
	if aIdx < 0 || bIdx < 0 {
		t.Fatalf("expected output blocks missing: %q", output)
	}
	if aIdx > bIdx {
		t.Fatalf("expected output to be sorted by relative path")
	}
	if clipboard != output {
		t.Fatalf("clipboard payload should match output content")
	}

	gitignoreBytes, err := os.ReadFile(filepath.Join(root, ".gitignore"))
	if err != nil {
		t.Fatalf("read .gitignore: %v", err)
	}
	if !strings.Contains(string(gitignoreBytes), "tools/\n") {
		t.Fatalf("expected tools/ entry in .gitignore")
	}
}

func TestExportSelectionRequiresSelectedFiles(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	mustWriteFile(t, filepath.Join(root, "a.go"), "package main\n")

	rootNode := &Node{
		Name:    filepath.Base(root),
		RelPath: ".",
		AbsPath: root,
		IsDir:   true,
		Children: []*Node{
			{
				Name:    "a.go",
				RelPath: "a.go",
				AbsPath: filepath.Join(root, "a.go"),
				Ext:     ".go",
			},
		},
	}

	_, err := ExportSelection(rootNode, ExportOptions{RootPath: root}, nil)
	if err == nil {
		t.Fatalf("expected error when no files are selected")
	}
}

func TestExportSelectionDoesNotRepatchExistingToolsIgnore(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	mustWriteFile(t, filepath.Join(root, "a.go"), "package main\n")
	mustWriteFile(t, filepath.Join(root, ".gitignore"), "node_modules/\ntools/\n")

	rootNode := &Node{
		Name:    filepath.Base(root),
		RelPath: ".",
		AbsPath: root,
		IsDir:   true,
	}
	selected := &Node{
		Name:     "a.go",
		RelPath:  "a.go",
		AbsPath:  filepath.Join(root, "a.go"),
		Ext:      ".go",
		Selected: true,
		Parent:   rootNode,
	}
	rootNode.Children = []*Node{selected}

	result, err := ExportSelection(rootNode, ExportOptions{RootPath: root}, nil)
	if err != nil {
		t.Fatalf("ExportSelection returned error: %v", err)
	}
	if result.GitIgnorePatched {
		t.Fatalf("expected existing tools/ ignore to avoid patching")
	}
}
