package core

import (
	"os"
	"path/filepath"
	"testing"
)

func TestResolveOutputPathDefault(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	got, err := ResolveOutputPath(root, "")
	if err != nil {
		t.Fatalf("ResolveOutputPath returned error: %v", err)
	}

	want := filepath.Join(root, "tools", "output.txt")
	if got != want {
		t.Fatalf("unexpected output path: got %q want %q", got, want)
	}
}

func TestResolveOutputPathDotUsesCurrentWorkingDirectory(t *testing.T) {
	root := t.TempDir()
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("get current working directory: %v", err)
	}
	t.Cleanup(func() {
		_ = os.Chdir(cwd)
	})

	workDir := t.TempDir()
	if err := os.Chdir(workDir); err != nil {
		t.Fatalf("change working directory: %v", err)
	}

	got, err := ResolveOutputPath(root, ".")
	if err != nil {
		t.Fatalf("ResolveOutputPath returned error: %v", err)
	}

	want := filepath.Join(workDir, "tools", "output.txt")
	if got != want {
		t.Fatalf("unexpected output path for dot spec: got %q want %q", got, want)
	}
}

func TestResolveOutputPathHandlesDirectoryAndFileTargets(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	dirTarget := filepath.Join(root, "artifacts")
	if err := os.MkdirAll(dirTarget, 0o755); err != nil {
		t.Fatalf("create directory target: %v", err)
	}

	gotDir, err := ResolveOutputPath(root, dirTarget)
	if err != nil {
		t.Fatalf("resolve existing directory target: %v", err)
	}
	wantDir := filepath.Join(dirTarget, "tools", "output.txt")
	if gotDir != wantDir {
		t.Fatalf("unexpected directory target resolution: got %q want %q", gotDir, wantDir)
	}

	fileTarget := filepath.Join(root, "merged.md")
	gotFile, err := ResolveOutputPath(root, fileTarget)
	if err != nil {
		t.Fatalf("resolve file target: %v", err)
	}
	wantFile, err := filepath.Abs(fileTarget)
	if err != nil {
		t.Fatalf("resolve absolute file target: %v", err)
	}
	if gotFile != wantFile {
		t.Fatalf("unexpected file target resolution: got %q want %q", gotFile, wantFile)
	}
}
