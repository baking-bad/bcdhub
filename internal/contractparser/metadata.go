package contractparser

import (
	"fmt"

	"github.com/tidwall/gjson"
)

// Metadata -
type Metadata map[string]*NodeMetadata

// NodeMetadata -
type NodeMetadata struct {
	TypeName  string   `json:"type,omitempty"`
	FieldName string   `json:"field,omitempty"`
	Prim      string   `json:"prim,omitempty"`
	Entry     string   `json:"entry,omitempty"`
	Parameter string   `json:"parameter,omitempty"`
	Args      []string `json:"args,omitempty"`
}

type internalNode struct {
	*Node
	InternalArgs []internalNode `json:"-,omitempty"`
	Nested       bool           `json:"-"`
}

func getAnnotation(x []string, prefix byte) string {
	for i := range x {
		if x[i][0] == prefix {
			return x[i][1:]
		}
	}
	return ""
}

// ParseMetadata -
func ParseMetadata(v gjson.Result) (Metadata, error) {
	m := make(Metadata)
	parent := Node{
		Prim: "root",
		Path: "0",
	}

	if v.IsArray() {
		val := v.Array()
		if len(val) > 0 {
			parseNodeMetadata(val[0], parent, parent.Path, "", m)
			return m, nil
		}
		return nil, fmt.Errorf("[ParseMetadata] Invalid data length: %d", len(val))
	} else if v.IsObject() {
		parseNodeMetadata(v, parent, parent.Path, "", m)
		return m, nil
	} else {
		return nil, fmt.Errorf("Unknown value type: %T", v.Type)
	}
}

func getFlatNested(data internalNode) []internalNode {
	nodes := make([]internalNode, 0)
	for _, arg := range data.InternalArgs {
		if data.Node.Is(arg.Node.Prim) && len(arg.InternalArgs) > 0 && arg.Nested {
			nodes = append(nodes, getFlatNested(arg)...)
		} else {
			nodes = append(nodes, arg)
		}
	}
	return nodes
}

func parseNodeMetadata(v gjson.Result, parent Node, path, entry string, metadata Metadata) internalNode {
	n := newNodeJSON(v)
	arr := n.Args.Array()
	n.Path = path

	fieldName := getAnnotation(n.Annotations, '%')
	typeName := getAnnotation(n.Annotations, ':')

	if _, ok := metadata[path]; !ok {
		metadata[path] = &NodeMetadata{
			Prim:      n.Prim,
			TypeName:  typeName,
			FieldName: fieldName,
			Entry:     entry,
			Args:      make([]string, 0),
		}
	}

	if n.Is(LAMBDA) || n.Is(CONTRACT) {
		if len(arr) > 0 {
			m := metadata[path]
			m.Parameter = arr[0].String()
		}
		return internalNode{
			Node: &n,
		}
	} else if n.Is(OPTION) {
		return parseNodeMetadata(arr[0], parent, path+"0", fieldName, metadata)
	}

	args := make([]internalNode, 0)
	for i := range arr {
		argPath := fmt.Sprintf("%s%d", path, i)
		args = append(args, parseNodeMetadata(arr[i], n, argPath, entry, metadata))
	}

	if n.Is(PAIR) || n.Is(OR) {
		res := internalNode{
			Node:         &n,
			InternalArgs: args,
			Nested:       true,
		}
		isStruct := n.Is(PAIR) && (typeName != "" || fieldName != "")
		if isStruct || n.Prim != parent.Prim {
			args = getFlatNested(res)
		} else {
			return res
		}
	}

	m := metadata[path]
	for _, a := range args {
		m.Args = append(m.Args, a.Node.Path)
	}

	return internalNode{
		Node:         &n,
		InternalArgs: args,
	}
}

// GetMetadataNetwork -
func GetMetadataNetwork(level int64) string {
	if level >= LevelBabylon {
		return "babylon"
	}
	return "alpha"
}
