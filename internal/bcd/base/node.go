package base

import (
	"fmt"
	"strings"

	"github.com/baking-bad/bcdhub/internal/bcd/consts"
	jsoniter "github.com/json-iterator/go"
)

var json = jsoniter.ConfigCompatibleWithStandardLibrary

// Node -
type Node struct {
	Prim        string   `json:"prim,omitempty"`
	Args        []*Node  `json:"args,omitempty"`
	Annots      []string `json:"annots,omitempty"`
	StringValue *string  `json:"string,omitempty"`
	BytesValue  *string  `json:"bytes,omitempty"`
	IntValue    *BigInt  `json:"int,omitempty"`
}

// UnmarshalJSON -
func (node *Node) UnmarshalJSON(data []byte) error {
	if len(data) == 0 {
		return consts.ErrInvalidJSON
	}
	if data[0] == '[' {
		node.Prim = consts.PrimArray
		node.Args = make([]*Node, 0)
		return json.Unmarshal(data, &node.Args)
	} else if data[0] == '{' {
		type buf Node
		return json.Unmarshal(data, (*buf)(node))
	}
	return consts.ErrInvalidJSON
}

// GetAnnotations -
func (node *Node) GetAnnotations() map[string]struct{} {
	annots := make(map[string]struct{})
	for i := range node.Annots {
		if len(node.Annots[i]) == 0 {
			continue
		}
		if node.Annots[i][0] == consts.AnnotPrefixFieldName || node.Annots[i][0] == consts.AnnotPrefixrefixTypeName {
			annots[node.Annots[i][1:]] = struct{}{}
		}
	}
	for i := range node.Args {
		for k := range node.Args[i].GetAnnotations() {
			annots[k] = struct{}{}
		}
	}
	return annots
}

// Hash -
func (node *Node) Hash() (string, error) {
	var s strings.Builder
	var prim string
	switch {
	case node.Prim != "":
		if node.Prim != consts.RENAME && node.Prim != consts.CAST && node.Prim != consts.PrimArray {
			hashCode, err := getHashCode(node.Prim)
			if err != nil {
				return "", err
			}
			s.WriteString(hashCode)
		}

		for i := range node.Args {
			childHashCode, err := node.Args[i].Hash()
			if err != nil {
				return "", err
			}
			s.WriteString(childHashCode)
		}
		return s.String(), nil
	case node.BytesValue != nil:
		prim = consts.BYTES
	case node.IntValue != nil:
		prim = consts.INT
	case node.StringValue != nil:
		prim = consts.STRING
	}
	hashCode, err := getHashCode(prim)
	if err != nil {
		return "", err
	}
	s.WriteString(hashCode)
	return s.String(), nil
}

// String -
func (node *Node) String() string {
	return node.print(0) + "\n"
}

func (node *Node) print(depth int) string {
	var s strings.Builder
	s.WriteByte('\n')
	s.WriteString(strings.Repeat(consts.DefaultIndent, depth))
	switch {
	case node.Prim != "":
		s.WriteString(node.Prim)
		for i := range node.Args {
			s.WriteString(node.Args[i].print(depth + 1))
		}
	case node.IntValue != nil:
		s.WriteString(fmt.Sprintf("Int=%d", *node.IntValue))
	case node.BytesValue != nil:
		s.WriteString(fmt.Sprintf("Bytes=%s", *node.BytesValue))
	case node.StringValue != nil:
		s.WriteString(fmt.Sprintf("String=%s", *node.StringValue))
	}
	return s.String()
}
