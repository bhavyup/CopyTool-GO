package cli

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"copytool/internal/core"
	"copytool/internal/platform"
)

type Options struct {
	RootPath    string
	Paths       []string
	Filters     core.FilterConfig
	OutputSpec  string
	ToClipboard bool
}

func RunExport(opts Options) error {
	root, err := filepath.Abs(opts.RootPath)
	if err != nil {
		return fmt.Errorf("resolve root path: %w", err)
	}

	paths, source, err := resolveInputPaths(root, opts.Paths)
	if err != nil {
		return err
	}

	tree, err := core.ScanRoot(root, opts.Filters)
	if err != nil {
		return fmt.Errorf("scan project tree: %w", err)
	}

	applied, skipped := core.ApplyInitialSelections(tree.Root, root, paths)
	selectedCount := core.CountSelectedFiles(tree.Root)

	if selectedCount == 0 {
		return fmt.Errorf("no eligible files selected after applying .gitignore and active filters")
	}

	resolvedOutputPath, err := core.ResolveOutputPath(root, opts.OutputSpec)
	if err != nil {
		return err
	}

	var clipboardWriter core.ClipboardWriter
	if opts.ToClipboard {
		clipboardWriter = platform.WriteClipboard
	}

	result, err := core.ExportSelection(tree.Root, core.ExportOptions{
		RootPath:    root,
		OutputSpec:  opts.OutputSpec,
		ToClipboard: opts.ToClipboard,
	}, clipboardWriter)
	if err != nil {
		return err
	}

	printSummary(root, source, opts.Filters, len(paths), applied, skipped, resolvedOutputPath, result)
	return nil
}

func resolveInputPaths(root string, explicit []string) ([]string, string, error) {
	if len(explicit) > 0 {
		return explicit, "command arguments", nil
	}

	listPath, err := ensureListFile(root)
	if err != nil {
		return nil, "", err
	}

	paths, err := readListFile(listPath)
	if err != nil {
		return nil, "", fmt.Errorf("read tools/list.txt: %w", err)
	}

	if len(paths) == 0 {
		return nil, "", fmt.Errorf("no paths passed and %s is empty; add entries or pass explicit paths", listPath)
	}

	return paths, "tools/list.txt", nil
}

func ensureListFile(root string) (string, error) {
	toolsDir := filepath.Join(root, "tools")
	if err := os.MkdirAll(toolsDir, 0755); err != nil {
		return "", fmt.Errorf("create tools directory: %w", err)
	}

	listPath := filepath.Join(toolsDir, "list.txt")
	if _, err := os.Stat(listPath); os.IsNotExist(err) {
		if err := os.WriteFile(listPath, []byte(""), 0644); err != nil {
			return "", fmt.Errorf("create tools/list.txt: %w", err)
		}
	} else if err != nil {
		return "", err
	}

	return listPath, nil
}

func readListFile(path string) ([]string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var paths []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		if strings.HasPrefix(line, "#") {
			continue
		}
		paths = append(paths, line)
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return paths, nil
}

func printSummary(
	root string,
	source string,
	filters core.FilterConfig,
	inputCount int,
	applied int,
	skipped []string,
	resolvedOutputPath string,
	result core.ExportResult,
) {
	fmt.Printf("Root: %s\n", root)
	fmt.Printf("Mode: CLI export\n")
	fmt.Printf("Selection source: %s\n", source)

	if len(filters.IncludeExts) > 0 {
		fmt.Printf("Include extensions: %s\n", core.FormatSetCSV(filters.IncludeExts))
	}
	if len(filters.ExcludeExts) > 0 {
		fmt.Printf("Exclude extensions: %s\n", core.FormatSetCSV(filters.ExcludeExts))
	}
	if len(filters.ExcludeDirs) > 0 {
		fmt.Printf("Exclude directories: %s\n", core.FormatSetCSV(filters.ExcludeDirs))
	}

	fmt.Printf("Resolved output: %s\n", resolvedOutputPath)
	fmt.Printf("Input paths: %d\n", inputCount)
	fmt.Printf("Applied selections: %d\n", applied)

	if len(skipped) > 0 {
		sort.Strings(skipped)
		fmt.Printf("Skipped selections: %d\n", len(skipped))
		fmt.Printf("Skipped list: %s\n", summarizeList(skipped, 8))
	}

	fmt.Printf("Exported files: %d/%d\n", result.ExportedCount, result.SelectedCount)

	if result.SkippedCount > 0 {
		fmt.Printf("Unreadable files skipped during export: %d\n", result.SkippedCount)
	}

	fmt.Printf("List file: %s\n", result.ListPath)
	fmt.Printf("Output file: %s\n", result.OutputPath)

	if result.Clipboard {
		fmt.Printf("Clipboard: updated\n")
	}

	if result.GitIgnorePatched {
		fmt.Printf(".gitignore: added tools/\n")
	}
}

func summarizeList(values []string, limit int) string {
	if len(values) <= limit {
		return strings.Join(values, ", ")
	}

	return strings.Join(values[:limit], ", ") + fmt.Sprintf(" +%d more", len(values)-limit)
}