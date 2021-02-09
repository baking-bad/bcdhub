package ast

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/baking-bad/bcdhub/internal/bcd/consts"
)

// MiguelNode -
type MiguelNode struct {
	Prim     string      `json:"prim,omitempty"`
	Type     string      `json:"type,omitempty"`
	Name     *string     `json:"name,omitempty"`
	From     interface{} `json:"from,omitempty"`
	DiffType string      `json:"diff_type,omitempty"`
	Value    interface{} `json:"value,omitempty"`

	Children []*MiguelNode `json:"children,omitempty"`
}

// String -
func (node *MiguelNode) String() string {
	return node.print(0)
}

func (node *MiguelNode) print(depth int) string {
	var s strings.Builder
	s.WriteString(strings.Repeat(consts.DefaultIndent, depth))
	if node.Value == nil {
		s.WriteString("option=nil")
	} else {
		s.WriteString(fmt.Sprintf("prim=%s", node.Prim))
	}
	if node.Name != nil {
		s.WriteString(fmt.Sprintf(" name=%s", *node.Name))
	}
	if node.Type != "" {
		s.WriteString(fmt.Sprintf(" type=%s", node.Type))
	}
	if node.Value != nil {
		s.WriteString(fmt.Sprintf(" value=%v", node.Value))
	}
	s.WriteByte('\n')
	for i := range node.Children {
		s.WriteString(node.Children[i].print(depth + 1))
	}
	return s.String()
}

// Compare -
func (node *MiguelNode) Compare(second *MiguelNode) bool {
	if second == nil {
		return false
	}
	if node.Prim != second.Prim || node.Type != second.Type {
		return false
	}
	if !isInterfaceEqual(node.Name, second.Name) {
		return false
	}
	if !isInterfaceEqual(node.Value, second.Value) {
		return false
	}
	if len(node.Children) != len(second.Children) {
		return false
	}
	for i := 0; i < len(node.Children); i++ {
		if !node.Children[i].Compare(second.Children[i]) {
			return false
		}
	}
	return true
}

func isInterfaceEqual(x, y interface{}) bool {
	switch {
	case x == nil && y == nil:
		return true
	case x != nil && y == nil:
		return false
	case x == nil && y != nil:
		return false
	default:
		return reflect.DeepEqual(x, y)
	}
}
