package core

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

type ClipboardWriter func(string) error

type ExportOptions struct {
	RootPath    string
	OutputSpec  string
	ToClipboard bool
}

type ExportResult struct {
	SelectedCount    int
	ExportedCount    int
	SkippedCount     int
	ListPath         string
	OutputPath       string
	Clipboard        bool
	GitIgnorePatched bool
}

func ExportSelection(root *Node, options ExportOptions, writeClipboard ClipboardWriter) (ExportResult, error) {
	if root == nil {
		return ExportResult{}, fmt.Errorf("tree root is nil")
	}

	selectedFiles := CollectSelectedFiles(root)
	if len(selectedFiles) == 0 {
		return ExportResult{}, fmt.Errorf("no files selected")
	}

	sort.Slice(selectedFiles, func(i, j int) bool {
		return filepath.ToSlash(selectedFiles[i].RelPath) < filepath.ToSlash(selectedFiles[j].RelPath)
	})

	rootToolsDir := filepath.Join(options.RootPath, "tools")
	listPath := filepath.Join(rootToolsDir, "list.txt")

	outputPath, err := ResolveOutputPath(options.RootPath, options.OutputSpec)
	if err != nil {
		return ExportResult{}, err
	}

	if err := os.MkdirAll(rootToolsDir, 0755); err != nil {
		return ExportResult{}, fmt.Errorf("create root tools directory: %w", err)
	}

	outputDir := filepath.Dir(outputPath)
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return ExportResult{}, fmt.Errorf("create output directory: %w", err)
	}

	gitIgnorePatched, err := ensureToolsIgnored(options.RootPath)
	if err != nil {
		return ExportResult{}, fmt.Errorf("update .gitignore: %w", err)
	}

	relPaths := make([]string, 0, len(selectedFiles))
	for _, node := range selectedFiles {
		relPaths = append(relPaths, filepath.ToSlash(node.RelPath))
	}

	listContent := ""
	if len(relPaths) > 0 {
		listContent = strings.Join(relPaths, "\n") + "\n"
	}

	if err := os.WriteFile(listPath, []byte(listContent), 0644); err != nil {
		return ExportResult{}, fmt.Errorf("write list.txt: %w", err)
	}

	var builder strings.Builder
	exportedCount := 0
	skippedCount := 0

	for _, node := range selectedFiles {
		data, err := os.ReadFile(node.AbsPath)
		if err != nil {
			skippedCount++
			continue
		}

		relPath := filepath.ToSlash(node.RelPath)
		lang := fenceLanguage(node.AbsPath)
		fence := "````"
		openFence := fence
		if lang != "" {
			openFence = fence + lang
		}

		builder.WriteString(relPath)
		builder.WriteString("\n")
		builder.WriteString(openFence)
		builder.WriteString("\n")
		builder.Write(data)

		if len(data) == 0 || data[len(data)-1] != '\n' {
			builder.WriteString("\n")
		}

		builder.WriteString(fence)
		builder.WriteString("\n\n")

		exportedCount++
	}

	outputContent := builder.String()

	if err := os.WriteFile(outputPath, []byte(outputContent), 0644); err != nil {
		return ExportResult{}, fmt.Errorf("write output file: %w", err)
	}

	if options.ToClipboard && writeClipboard != nil {
		if err := writeClipboard(outputContent); err != nil {
			return ExportResult{}, fmt.Errorf("write clipboard: %w", err)
		}
	}

	return ExportResult{
		SelectedCount:    len(selectedFiles),
		ExportedCount:    exportedCount,
		SkippedCount:     skippedCount,
		ListPath:         listPath,
		OutputPath:       outputPath,
		Clipboard:        options.ToClipboard,
		GitIgnorePatched: gitIgnorePatched,
	}, nil
}

func CollectSelectedFiles(root *Node) []*Node {
	var files []*Node

	var walk func(*Node)
	walk = func(node *Node) {
		if node == nil {
			return
		}

		if node.IsDir {
			for _, child := range node.Children {
				walk(child)
			}
			return
		}

		if node.Selected {
			files = append(files, node)
		}
	}

	walk(root)
	return files
}

func ensureToolsIgnored(rootPath string) (bool, error) {
	gitignorePath := filepath.Join(rootPath, ".gitignore")

	data, err := os.ReadFile(gitignorePath)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, err
	}

	content := string(data)
	normalized := strings.ReplaceAll(content, "\r\n", "\n")
	lines := strings.Split(normalized, "\n")

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		switch trimmed {
		case "tools/", "/tools/", "tools", "/tools":
			return false, nil
		}
	}

	if len(content) > 0 && !strings.HasSuffix(content, "\n") && !strings.HasSuffix(content, "\r\n") {
		content += "\n"
	}
	content += "tools/\n"

	if err := os.WriteFile(gitignorePath, []byte(content), 0644); err != nil {
		return false, err
	}

	return true, nil
}

func fenceLanguage(path string) string {
	switch strings.ToLower(filepath.Ext(path)) {
	case ".ps1", ".psm1":
		return "powershell"
	case ".cmd", ".bat":
		return "bat"
	case ".cs":
		return "csharp"
	case ".js":
		return "javascript"
	case ".jsx":
		return "jsx"
	case ".ts":
		return "typescript"
	case ".tsx":
		return "tsx"
	case ".json":
		return "json"
	case ".md":
		return "markdown"
	case ".html":
		return "html"
	case ".css":
		return "css"
	case ".scss":
		return "scss"
	case ".py":
		return "python"
	case ".java":
		return "java"
	case ".cpp", ".cxx", ".cc", ".hpp":
		return "cpp"
	case ".c", ".h":
		return "c"
	case ".xml":
		return "xml"
	case ".yml", ".yaml":
		return "yaml"
	case ".sql":
		return "sql"
	case ".sh":
		return "bash"
	case ".go":
		return "go"
	case ".rs":
		return "rust"
	case ".php":
		return "php"
	case ".rb":
		return "ruby"
	case ".kt":
		return "kotlin"
	case ".swift":
		return "swift"
	default:
		return ""
	}
}
