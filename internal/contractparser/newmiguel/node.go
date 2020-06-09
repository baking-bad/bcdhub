package newmiguel

import (
	"reflect"

	"github.com/baking-bad/bcdhub/internal/contractparser/consts"
	"github.com/baking-bad/bcdhub/internal/helpers"
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
	IsOption bool        `json:"-"`

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
		if !helpers.StringInArray(node.Type, []string{
			consts.BIGMAP, consts.MAP, consts.LIST, consts.SET, consts.TypeNamedTuple, consts.TypeNamedUnion,
		}) {
			node.setDiffType(diffType)
			return
		}
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
	case consts.LIST, consts.SET:
		listMerge(node, second)
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
				node.Children[i].setDiffType(delete)
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

	removed := true
	for i := range node.Children {
		if node.Children[i].DiffType != delete {
			removed = false
			break
		}
	}

	if removed {
		node.DiffType = delete
	}
}

func (node *Node) compare(second *Node) bool {
	if !node.compareFields(second) {
		return false
	}
	if !node.compareValue(second) {
		return false
	}
	if len(node.Children) != len(second.Children) {
		return false
	}
	for i := 0; i < len(node.Children); i++ {
		if !node.Children[i].compare(second.Children[i]) {
			return false
		}
	}
	return true
}

func getMatrix(first, second []*Node) [][]int {
	n := len(second)
	m := len(first)

	d := make([][]int, m+1)
	for i := 0; i < m+1; i++ {
		d[i] = make([]int, n+1, n+1)
		d[i][0] = i
	}
	for j := range d[0] {
		d[0][j] = j
	}

	for j := 1; j < n+1; j++ {
		for i := 1; i < m+1; i++ {
			cost := 0
			if !first[i-1].compare(second[j-1]) {
				cost = 1
			}

			d[i][j] = min(min(d[i-1][j]+1, d[i][j-1]+1), d[i-1][j-1]+cost)
		}
	}
	return d
}

func mergeMatrix(d [][]int, i, j int, first, second *Node) []*Node {
	children := make([]*Node, 0)
	if i == 0 && j == 0 {
		return children
	}
	if i == 0 {
		for idx := 0; idx < j; idx++ {
			second.Children[idx].setDiffType(delete)
			children = append(children, second.Children[idx])
		}
		return children
	}
	if j == 0 {
		for idx := 0; idx < i; idx++ {
			first.Children[idx].setDiffType(create)
			children = append(children, first.Children[idx])
		}
		return children
	}
	left := d[i][j-1]
	up := d[i-1][j]
	upleft := d[i-1][j-1]

	if upleft <= up && upleft <= left {
		if upleft == d[i][j] {
			children = mergeMatrix(d, i-1, j-1, first, second)
			children = append(children, first.Children[i-1])
		} else {
			children = mergeMatrix(d, i-1, j-1, first, second)
			first.Children[i-1].setDiffType(update)
			first.Children[i-1].From = second.Children[j-1].Value
			children = append(children, first.Children[i-1])
		}
	} else {
		if left <= upleft && left <= up {
			children = mergeMatrix(d, i, j-1, first, second)
			second.Children[j-1].setDiffType(delete)
			children = append(children, second.Children[j-1])
		} else {
			children = mergeMatrix(d, i-1, j, first, second)
			first.Children[i-1].setDiffType(create)
			children = append(children, first.Children[i-1])
		}
	}
	return children
}

func listMerge(first, second *Node) {
	d := getMatrix(first.Children, second.Children)
	children := mergeMatrix(d, len(first.Children), len(second.Children), first, second)
	first.Children = children
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
				if node.Children[i].Value == nil {
					node.Children[i].Value = second.Children[j].Value
					node.Children[i].setDiffType(delete)
				} else if second.Children[j].Value == nil {
					node.Children[i].setDiffType(create)
				} else {
					node.Children[i].setDiffType(update)
					node.Children[i].From = second.Children[j].Value
				}
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
