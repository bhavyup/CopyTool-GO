package core

type Node struct {
	Name     string
	RelPath  string
	AbsPath  string
	IsDir    bool
	Parent   *Node
	Children []*Node
	Expanded bool
	Size     int64
	Ext      string

	Selected bool
	Partial  bool
}

func (n *Node) Depth() int {
	depth := 0
	for current := n.Parent; current != nil; current = current.Parent {
		depth++
	}
	return depth
}

type Stats struct {
	Nodes int
	Files int
	Dirs  int
}

func ComputeStats(node *Node) Stats {
	var stats Stats

	var walk func(*Node)
	walk = func(n *Node) {
		if n == nil {
			return
		}

		stats.Nodes++
		if n.IsDir {
			stats.Dirs++
		} else {
			stats.Files++
		}

		for _, child := range n.Children {
			walk(child)
		}
	}

	walk(node)
	return stats
}