package tui

import (
	"sort"
	"strings"

	"copytool/internal/core"

	"github.com/sahilm/fuzzy"
)

type paletteItem struct {
	RelPath string
	IsDir   bool
	Label   string // rendered label, also used for matching
}

func (m *Model) rebuildPaletteIndex() {
	m.PaletteAll = nil
	m.PaletteResults = nil
	m.PaletteIndex = 0

	if m.Tree == nil || m.Tree.Root == nil {
		return
	}

	var items []paletteItem
	var walk func(*core.Node)
	walk = func(n *core.Node) {
		if n == nil {
			return
		}

		// Skip root "." entry; it's not helpful in search.
		if n.RelPath != "." {
			label := strings.ReplaceAll(n.RelPath, `\`, `/`)
			if n.IsDir {
				label += "/"
			}
			items = append(items, paletteItem{
				RelPath: n.RelPath,
				IsDir:   n.IsDir,
				Label:   label,
			})
		}

		if n.IsDir {
			for _, c := range n.Children {
				walk(c)
			}
		}
	}

	walk(m.Tree.Root)

	// Stable sort by label so search results feel consistent
	sort.Slice(items, func(i, j int) bool {
		return strings.ToLower(items[i].Label) < strings.ToLower(items[j].Label)
	})

	m.PaletteAll = items
}

func (m *Model) paletteRefreshResults() {
	query := strings.TrimSpace(m.PaletteInput.Value())
	if query == "" {
		// For empty query, show a small "top slice" so palette isn't blank.
		m.PaletteResults = takeFirst(m.PaletteAll, 40)
		m.PaletteIndex = 0
		return
	}

	candidates := make([]string, 0, len(m.PaletteAll))
	for _, item := range m.PaletteAll {
		candidates = append(candidates, item.Label)
	}

	matches := fuzzy.Find(query, candidates)
	results := make([]paletteItem, 0, min(len(matches), 80))
	for i := 0; i < len(matches) && i < 80; i++ {
		results = append(results, m.PaletteAll[matches[i].Index])
	}

	m.PaletteResults = results
	m.PaletteIndex = 0
}

func takeFirst[T any](items []T, n int) []T {
	if n <= 0 {
		return nil
	}
	if len(items) <= n {
		return items
	}
	return items[:n]
}
