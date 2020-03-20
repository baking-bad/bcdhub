package newmiguel

import (
	"reflect"

	"github.com/baking-bad/bcdhub/internal/contractparser/consts"
)

const (
	create = "create"
	update = "update"
	delete = "delete"
)

// Node -
type Node struct {
	Prim     string      `json:"prim,omitempty"`
	Type     string      `json:"type,omitempty"`
	Name     string      `json:"name,omitempty"`
	From     interface{} `json:"from,omitempty"`
	DiffType string      `json:"diff_type,omitempty"`
	Value    interface{} `json:"value,omitempty"`

	Children []*Node `json:"children,omitempty"`
}

// Diff -
func (node *Node) Diff(prev *Node) {
	if prev == nil {
		node.setDiffType(create)
		return
	}
	if !node.compareFields(prev) {
		node.setDiffType(create)
		return
	}

	if !node.compareValue(prev) {
		if prev.Value == nil {
			node.setDiffType(create)
			return
		} else if node.Value == nil {
			node = prev
			node.setDiffType(delete)
			return
		} else {
			node.DiffType = update
			node.From = prev.Value
		}
	}
	node.compareChildren(prev)
}

func (node *Node) compareFields(second *Node) bool {
	if second == nil {
		return false
	}

	if node.Prim != second.Prim {
		return false
	}
	if node.Type != second.Type {
		return false
	}
	if node.Name != second.Name {
		return false
	}
	return true
}

func (node *Node) compareValue(second *Node) bool {
	if second == nil {
		return false
	}
	if second.Value == nil && node.Value == nil {
		return true
	}
	if node.Type == consts.BIGMAP && second.Type == consts.BIGMAP {
		return true
	}
	return reflect.DeepEqual(node.Value, second.Value)
}

func (node *Node) compareChildren(second *Node) {
	if len(node.Children) == 0 && len(second.Children) == 0 {
		return
	}

	var diffType string
	if len(node.Children) == 0 && len(second.Children) != 0 {
		diffType = delete
	} else if len(node.Children) != 0 && len(second.Children) == 0 {
		diffType = create
	}
	if diffType != "" {
		if node.Type != consts.MAP && node.Type != consts.BIGMAP {
			node.setDiffType(diffType)
		} else {
			for i := range node.Children {
				node.Children[i].setDiffType(diffType)
			}
		}
		return
	}

	merge(node, second)
}

func (node *Node) setDiffType(typ string) {
	node.DiffType = typ
	for i := range node.Children {
		if node.Children[i] == nil {
			node.Children[i] = &Node{}
		}
		node.Children[i].setDiffType(typ)
	}
}

func merge(node, second *Node) {
	switch node.Type {
	case consts.BIGMAP, consts.MAP:
		mapMerge(node, second)
	default:
		defaultMerge(node, second)
	}
}

func defaultMerge(node, second *Node) {
	var j int
	for i := 0; i < len(node.Children) && j < len(second.Children); i, j = i+1, j+1 {
		if node.Children[i] == nil {
			if second.Children[j] != nil {
				node.Children[i] = second.Children[j]
				node.Children[i].setDiffType(create)
			} else {
				node.Children[i] = &Node{}
			}
			continue
		}
		if second.Children[i] == nil {
			node.Children[i].setDiffType(create)
			continue
		}
		if !node.Children[i].compareFields(second.Children[j]) {
			if node.Children[i].Prim == "" {
				node.Children[i] = second.Children[j]
				node.Children[i].setDiffType(delete)
				continue
			}
			if second.Children[i].Prim == "" {
				node.Children[i].setDiffType(create)
				continue
			}
			i--
			continue
		}
		node.Children[i].Diff(second.Children[j])
	}

	if len(node.Children) > j {
		for i := j; i < len(node.Children); i++ {
			node.Children[i].setDiffType(create)
		}
	} else if len(second.Children) > j {
		for i := j; i < len(second.Children); i++ {
			second.Children[i].setDiffType(delete)
			node.Children = append(node.Children, second.Children[i])
		}
	}
}

func mapMerge(node, second *Node) {
	count := 0
	for i := range node.Children {
		found := false
		for j := range second.Children {
			if node.Children[i].Name != second.Children[j].Name {
				continue
			}
			if !node.Children[i].compareValue(second.Children[j]) {
				node.Children[i].setDiffType(update)
				node.Children[i].From = second.Children[j].Value
			} else {
				node.Children[i].compareChildren(second.Children[j])
			}
			found = true
			count++
		}

		if !found {
			node.Children[i].setDiffType(create)
		}
	}

	if count < len(second.Children) {
		for j := range second.Children {
			found := false
			for i := range node.Children {
				if !node.Children[i].compareFields(second.Children[j]) {
					continue
				}
				found = true
			}
			if !found && second.Children[j].DiffType == "" {
				second.Children[j].setDiffType(delete)
				node.Children = append(node.Children, second.Children[j])
			}
		}
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
