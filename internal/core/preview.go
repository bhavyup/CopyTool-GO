package core

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"strings"
	"unicode/utf8"
)

const (
	previewByteLimit  = 64 * 1024
	previewLineLimit  = 32
	folderSampleLimit = 10
)

type PreviewData struct {
	Title    string
	Subtitle string
	Lines    []string
	Numbered bool
	Notice   string
}

func BuildPreview(node *Node) PreviewData {
	if node == nil {
		return PreviewData{
			Title:  "No focus",
			Notice: "Nothing is currently focused.",
		}
	}

	if node.IsDir {
		return buildFolderPreview(node)
	}

	return buildFilePreview(node)
}

func buildFolderPreview(node *Node) PreviewData {
	stats := ComputeStats(node)
	selectedFiles := CountSelectedFiles(node)

	lines := []string{
		"subtree snapshot",
		"",
		fmt.Sprintf("selected files:   %d", selectedFiles),
		fmt.Sprintf("folders inside:   %d", maxInt(0, stats.Dirs-1)),
		fmt.Sprintf("files inside:     %d", stats.Files),
		fmt.Sprintf("children inside:  %d", len(node.Children)),
		"",
		"sample contents",
	}

	if len(node.Children) == 0 {
		lines = append(lines, "(empty folder)")
	} else {
		limit := minInt(folderSampleLimit, len(node.Children))
		for _, child := range node.Children[:limit] {
			label := child.Name
			if child.IsDir {
				label += "/"
			}

			prefix := "·"
			if child.IsDir {
				prefix = "▸"
			}

			lines = append(lines, fmt.Sprintf("%s %s %s", prefix, checkboxForPreview(child), label))
		}

		if len(node.Children) > limit {
			lines = append(lines, fmt.Sprintf("… %d more item(s)", len(node.Children)-limit))
		}
	}

	return PreviewData{
		Title:    node.Name,
		Subtitle: node.RelPath,
		Lines:    lines,
		Numbered: false,
	}
}

func buildFilePreview(node *Node) PreviewData {
	data, truncatedByBytes, err := readPreviewBytes(node.AbsPath, previewByteLimit)
	if err != nil {
		return PreviewData{
			Title:    node.Name,
			Subtitle: node.RelPath,
			Notice:   fmt.Sprintf("Unable to read file preview: %v", err),
		}
	}

	if len(data) == 0 {
		return PreviewData{
			Title:    node.Name,
			Subtitle: node.RelPath,
			Lines:    []string{"(empty file)"},
			Numbered: false,
		}
	}

	if bytes.IndexByte(data, 0) >= 0 {
		return PreviewData{
			Title:    node.Name,
			Subtitle: node.RelPath,
			Notice:   "Binary file detected. Literal preview skipped.",
		}
	}

	trimmed := trimToValidUTF8(data)
	if len(trimmed) == 0 && len(data) > 0 {
		return PreviewData{
			Title:    node.Name,
			Subtitle: node.RelPath,
			Notice:   "Non-UTF8 content detected. Literal preview skipped.",
		}
	}

	text := strings.ReplaceAll(string(trimmed), "\r\n", "\n")
	text = strings.ReplaceAll(text, "\r", "\n")

	allLines := strings.Split(text, "\n")
	if len(allLines) == 1 && allLines[0] == "" {
		allLines = []string{"(empty file)"}
	}

	truncatedByLines := false
	if len(allLines) > previewLineLimit {
		allLines = allLines[:previewLineLimit]
		truncatedByLines = true
	}

	for i, line := range allLines {
		allLines[i] = strings.ReplaceAll(line, "\t", "    ")
	}

	noticeParts := []string{}
	if truncatedByBytes {
		noticeParts = append(noticeParts, fmt.Sprintf("Preview source capped at first %d KiB.", previewByteLimit/1024))
	}
	if truncatedByLines {
		noticeParts = append(noticeParts, fmt.Sprintf("Preview buffer capped at first %d lines.", previewLineLimit))
	}

	return PreviewData{
		Title:    node.Name,
		Subtitle: node.RelPath,
		Lines:    allLines,
		Numbered: true,
		Notice:   strings.Join(noticeParts, " "),
	}
}

func readPreviewBytes(path string, limit int64) ([]byte, bool, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, false, err
	}
	defer file.Close()

	reader := io.LimitReader(file, limit+1)
	data, err := io.ReadAll(reader)
	if err != nil {
		return nil, false, err
	}

	truncated := int64(len(data)) > limit
	if truncated {
		data = data[:limit]
	}

	return data, truncated, nil
}

func trimToValidUTF8(data []byte) []byte {
	if utf8.Valid(data) {
		return data
	}

	trimmed := data
	for len(trimmed) > 0 && !utf8.Valid(trimmed) {
		trimmed = trimmed[:len(trimmed)-1]
	}

	return trimmed
}

func checkboxForPreview(node *Node) string {
	if node == nil {
		return "[ ]"
	}

	if node.Partial {
		return "[-]"
	}

	if node.Selected {
		return "[x]"
	}

	return "[ ]"
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}