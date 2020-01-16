package contractparser

import "fmt"

// Metadata -
type Metadata map[string]*NodeMetadata

// NodeMetadata -
type NodeMetadata struct {
	TypeName  string                 `json:"type,omitempty"`
	FieldName string                 `json:"field,omitempty"`
	Prim      string                 `json:"prim,omitempty"`
	Entry     string                 `json:"entry,omitempty"`
	Parameter map[string]interface{} `json:"parameter,omitempty"`
	Args      []string               `json:"args,omitempty"`
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
func ParseMetadata(v interface{}) (Metadata, error) {
	m := make(Metadata)
	parent := Node{
		Prim: "root",
		Path: "0",
	}

	switch val := v.(type) {
	case []interface{}:
		if len(val) > 0 {
			parseNodeMetadata(val[0].(map[string]interface{}), parent, parent.Path, "", m)
			return m, nil
		}
		return nil, fmt.Errorf("[ParseMetadata] Invalid data length: %d", len(val))
	case map[string]interface{}:
		parseNodeMetadata(val, parent, parent.Path, "", m)
		return m, nil
	default:
		return nil, fmt.Errorf("Unknown value type: %T", val)
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

func parseNodeMetadata(v map[string]interface{}, parent Node, path, entry string, metadata Metadata) internalNode {
	n := newNode(v)
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
		if len(n.Args) > 0 {
			arg := n.Args[0].(map[string]interface{})
			m := metadata[path]
			m.Parameter = arg
		}
		return internalNode{
			Node: &n,
		}
	} else if n.Is(OPTION) {
		arg := n.Args[0].(map[string]interface{})
		return parseNodeMetadata(arg, parent, path+"0", fieldName, metadata)
	}

	args := make([]internalNode, 0)
	for i := range n.Args {
		argPath := fmt.Sprintf("%s%d", path, i)
		a := n.Args[i].(map[string]interface{})
		args = append(args, parseNodeMetadata(a, n, argPath, entry, metadata))
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
