package contractparser

import (
	"fmt"
	"strings"

	"github.com/aopoltorzhicky/bcdhub/internal/helpers"
	"github.com/tidwall/gjson"
)

// Metadata -
type Metadata map[string]*NodeMetadata

// NodeMetadata -
type NodeMetadata struct {
	TypeName      string   `json:"typename,omitempty"`
	FieldName     string   `json:"fieldname,omitempty"`
	InheritedName string   `json:"-"`
	Prim          string   `json:"prim,omitempty"`
	Parameter     string   `json:"parameter,omitempty"`
	Args          []string `json:"args,omitempty"`
	Type          string   `json:"type,omitempty"`
	Name          string   `json:"name,omitempty"`
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

func parseNodeMetadata(v gjson.Result, parent Node, path, inheritedName string, metadata Metadata) internalNode {
	n := newNodeJSON(v)
	arr := n.Args.Array()
	n.Path = path

	fieldName := getAnnotation(n.Annotations, '%')
	typeName := getAnnotation(n.Annotations, ':')

	if _, ok := metadata[path]; !ok {
		metadata[path] = &NodeMetadata{
			Prim:          n.Prim,
			TypeName:      typeName,
			FieldName:     fieldName,
			InheritedName: inheritedName,
			Type:          n.Prim,
			Args:          make([]string, 0),
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
		return parseNodeMetadata(arr[0], parent, path+"/o", fieldName, metadata)
	}

	args := make([]internalNode, 0)
	if n.Is(MAP) || n.Is(BIGMAP) {
		if len(arr) == 2 {
			// parse key type
			args = append(args, parseNodeMetadata(arr[0], n, path+"/k", "", metadata))
			// parse value type
			args = append(args, parseNodeMetadata(arr[1], n, path+"/v", "", metadata))
			return internalNode{
				Node:         &n,
				InternalArgs: args,
			}
		}
	} else if n.Is(LIST) {
		if len(arr) == 1 {
			args = append(args, parseNodeMetadata(arr[0], n, path+"/l", "", metadata))
			return internalNode{
				Node:         &n,
				InternalArgs: args,
			}
		}
	} else if n.Is(SET) {
		if len(arr) == 1 {
			args = append(args, parseNodeMetadata(arr[0], n, path+"/s", "", metadata))
			return internalNode{
				Node:         &n,
				InternalArgs: args,
			}
		}
	} else {
		for i := range arr {
			argPath := fmt.Sprintf("%s/%d", path, i)
			args = append(args, parseNodeMetadata(arr[i], n, argPath, "", metadata))
		}

		if n.Is(PAIR) || n.Is(OR) {
			res := internalNode{
				Node:         &n,
				InternalArgs: args,
				Nested:       true,
			}

			isStruct := n.Is(PAIR) && (typeName != "" || fieldName != "" || inheritedName != "")
			if isStruct || n.Prim != parent.Prim {
				args = getFlatNested(res)
			} else {
				finishParseMetadata(metadata, path, res)
				return res
			}
		}
	}

	m := metadata[path]
	for _, a := range args {
		m.Args = append(m.Args, a.Node.Path)
	}

	in := internalNode{
		Node:         &n,
		InternalArgs: args,
	}
	finishParseMetadata(metadata, path, in)
	return in
}

func finishParseMetadata(metadata Metadata, path string, node internalNode) {
	if len(metadata[path].Args) > 0 {
		typ, keys := getNodeType(node, metadata)
		metadata[path].Type = typ
		for i := range keys {
			argPath := metadata[path].Args[i]
			metadata[argPath].Name = keys[i]
		}
	}
}

// GetMetadataNetwork -
func GetMetadataNetwork(level int64) string {
	if level >= LevelBabylon {
		return "babylon"
	}
	return "alpha"
}

func getKey(metadata *NodeMetadata) string {
	if metadata.TypeName != "" {
		return metadata.TypeName
	} else if metadata.FieldName != "" {
		return metadata.FieldName
	} else if metadata.InheritedName != "" {
		return metadata.InheritedName
	} else if helpers.StringInArray(metadata.Prim, []string{
		KEY, KEYHASH, SIGNATURE, TIMESTAMP, ADDRESS,
	}) {
		return metadata.Prim
	}
	return ""
}

func allArgsIsUnit(n internalNode, metadata Metadata) bool {
	nm := metadata[n.Path]
	for _, arg := range nm.Args {
		if metadata[arg].Prim != UNIT {
			return false
		}
	}
	return true
}

func getEntry(metadata *NodeMetadata) string {
	entry := ""
	if metadata.InheritedName != "" {
		entry = metadata.InheritedName
	} else if metadata.FieldName != "" {
		entry = metadata.FieldName
	} else if metadata.TypeName != "" {
		entry = metadata.TypeName
	}
	return strings.ReplaceAll(entry, "_Liq_entry_", "")
}

func getPairType(n internalNode, metadata Metadata) (string, []string) {
	nm := metadata[n.Path]

	keys := make([]string, 0)
	for _, arg := range nm.Args {
		m := metadata[arg]
		keys = append(keys, getKey(m))
	}
	if arrayUniqueLen(keys) == len(nm.Args) {
		return TypeNamedTuple, keys
	}
	return TypeTuple, nil
}

func getOrType(n internalNode, metadata Metadata) (string, []string) {
	nm := metadata[n.Path]

	entries := make([]string, 0)
	for _, arg := range nm.Args {
		m := metadata[arg]
		entries = append(entries, getEntry(m))
	}
	named := arrayUniqueLen(entries) == len(nm.Args)

	if allArgsIsUnit(n, metadata) {
		if named {
			return TypeNamedEnum, entries
		}
		return TypeEnum, nil
	}

	if named {
		return TypeNamedUnion, entries
	}
	return TypeUnion, nil
}

func getNodeType(n internalNode, metadata Metadata) (string, []string) {
	switch n.Prim {
	case OR:
		return getOrType(n, metadata)
	case PAIR:
		return getPairType(n, metadata)
	}
	return "", nil
}
