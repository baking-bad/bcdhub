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

	return reflect.DeepEqual(node.Value, second.Value)
}

func (node *Node) compareChildren(second *Node) {
	length := min(len(node.Children), len(second.Children))

	if length == 0 {
		if len(node.Children) == 0 && len(second.Children) != 0 {
			node.setDiffType(delete)
		} else if len(node.Children) != 0 && len(second.Children) == 0 {
			node.setDiffType(create)
		}
		return
	}

	j := 0
	for i := 0; i < length; i++ {
		if !node.Children[i].compareFields(second.Children[j]) {
			if node.Children[i].Prim == "" {
				node.Children[i] = second.Children[j]
				node.Children[i].setDiffType(delete)
				j++
				continue
			}
			if second.Children[i].Prim == "" {
				node.Children[i].setDiffType(create)
				j++
				continue
			}
			second.Children[j].setDiffType(create)
			i--
			j++
			continue
		}
		node.Children[i].Diff(second.Children[j])
		j++
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

	var diffType string
	eq := true
	for i := range node.Children {
		if i == 0 {
			diffType = node.Children[i].DiffType
		} else {
			if diffType != node.Children[i].DiffType {
				eq = false
				break
			}
		}
	}

	if eq && node.Type != consts.BIGMAP && node.Type != consts.MAP {
		node.DiffType = diffType
	}
}

func (node *Node) setDiffType(typ string) {
	node.DiffType = typ
	for i := range node.Children {
		node.Children[i].setDiffType(typ)
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
