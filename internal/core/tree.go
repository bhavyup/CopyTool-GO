package core

type Tree struct {
	Root         *Node
	VisibleNodes []*Node
	FocusIndex   int
}

func NewTree(root *Node) *Tree {
	tree := &Tree{
		Root:       root,
		FocusIndex: 0,
	}
	tree.RebuildVisible()
	return tree
}

func (t *Tree) RebuildVisible() {
	var visible []*Node

	var walk func(*Node)
	walk = func(node *Node) {
		if node == nil {
			return
		}

		visible = append(visible, node)

		if node.IsDir && node.Expanded {
			for _, child := range node.Children {
				walk(child)
			}
		}
	}

	walk(t.Root)
	t.VisibleNodes = visible

	if len(t.VisibleNodes) == 0 {
		t.FocusIndex = 0
		return
	}

	if t.FocusIndex < 0 {
		t.FocusIndex = 0
	}

	if t.FocusIndex >= len(t.VisibleNodes) {
		t.FocusIndex = len(t.VisibleNodes) - 1
	}
}

func (t *Tree) FocusedNode() *Node {
	if len(t.VisibleNodes) == 0 {
		return nil
	}

	if t.FocusIndex < 0 || t.FocusIndex >= len(t.VisibleNodes) {
		return t.VisibleNodes[0]
	}

	return t.VisibleNodes[t.FocusIndex]
}

func (t *Tree) FocusNode(target *Node) {
	if target == nil {
		return
	}

	for i, node := range t.VisibleNodes {
		if node == target {
			t.FocusIndex = i
			return
		}
	}
}

func (t *Tree) MoveUp() bool {
	if len(t.VisibleNodes) == 0 || t.FocusIndex <= 0 {
		return false
	}

	t.FocusIndex--
	return true
}

func (t *Tree) MoveDown() bool {
	if len(t.VisibleNodes) == 0 || t.FocusIndex >= len(t.VisibleNodes)-1 {
		return false
	}

	t.FocusIndex++
	return true
}

func (t *Tree) ToggleFocusedExpand() bool {
	node := t.FocusedNode()
	if node == nil || !node.IsDir || len(node.Children) == 0 {
		return false
	}

	node.Expanded = !node.Expanded
	t.RebuildVisible()
	t.FocusNode(node)
	return true
}

func (t *Tree) ExpandFocused() bool {
	node := t.FocusedNode()
	if node == nil || !node.IsDir || len(node.Children) == 0 || node.Expanded {
		return false
	}

	node.Expanded = true
	t.RebuildVisible()
	t.FocusNode(node)
	return true
}

func (t *Tree) CollapseFocused() bool {
	node := t.FocusedNode()
	if node == nil {
		return false
	}

	if node.IsDir && node.Expanded {
		node.Expanded = false
		t.RebuildVisible()
		t.FocusNode(node)
		return true
	}

	if node.Parent != nil {
		t.FocusNode(node.Parent)
		return true
	}

	return false
}