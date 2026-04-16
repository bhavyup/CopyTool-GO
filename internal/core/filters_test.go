package core

import "testing"

func TestParseExtensionCSVNormalizesValues(t *testing.T) {
	t.Parallel()

	got := ParseExtensionCSV("go, .MD, ., ,Go,.go")
	want := map[string]bool{
		".go": true,
		".md": true,
	}

	if len(got) != len(want) {
		t.Fatalf("unexpected set size: got %d want %d", len(got), len(want))
	}

	for ext := range want {
		if !got[ext] {
			t.Fatalf("expected extension %q to be present", ext)
		}
	}
}

func TestFilePassesExtensionFilters(t *testing.T) {
	t.Parallel()

	filters := FilterConfig{
		IncludeExts: ParseExtensionCSV("go,md"),
		ExcludeExts: ParseExtensionCSV("md"),
	}

	if !FilePassesExtensionFilters("main.go", filters) {
		t.Fatalf("expected .go file to pass include/exclude filters")
	}

	if FilePassesExtensionFilters("README.md", filters) {
		t.Fatalf("expected .md file to be excluded")
	}

	if FilePassesExtensionFilters("notes.txt", filters) {
		t.Fatalf("expected .txt file to be blocked by include filter")
	}
}

func TestDefaultExcludeDirsContainsHardExclusions(t *testing.T) {
	t.Parallel()

	dirs := DefaultExcludeDirs()
	for _, expected := range []string{".git", "tools", "node_modules", "dist"} {
		if !dirs[expected] {
			t.Fatalf("expected %q to be in default exclusion set", expected)
		}
	}

	if !IsHardExcludedDir("TOOLS") {
		t.Fatalf("hard exclusions should be case-insensitive")
	}
}
