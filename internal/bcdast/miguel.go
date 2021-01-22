package bcdast

import (
	"fmt"
	"strings"
)

// MiguelNode -
type MiguelNode struct {
	Prim     string      `json:"prim,omitempty"`
	Type     string      `json:"type,omitempty"`
	Name     string      `json:"name,omitempty"`
	From     interface{} `json:"from,omitempty"`
	DiffType string      `json:"diff_type,omitempty"`
	Value    interface{} `json:"value,omitempty"`
	IsOption bool        `json:"-"`

	Children []*MiguelNode `json:"children,omitempty"`
}

// String -
func (node *MiguelNode) String() string {
	return node.print(0)
}

func (node *MiguelNode) print(depth int) string {
	var s strings.Builder
	s.WriteString(strings.Repeat(indent, depth))
	if node.IsOption && node.Value == nil {
		s.WriteString("option=nil")
	} else {
		s.WriteString(fmt.Sprintf("prim=%s", node.Prim))
	}
	if node.Name != "" {
		s.WriteString(fmt.Sprintf(" name=%s", node.Name))
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
