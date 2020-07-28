package newmiguel

import (
	"strings"

	"github.com/baking-bad/bcdhub/internal/contractparser/meta"
	"github.com/tidwall/gjson"
)

type unionDecoder struct {
	parent *miguel
}

// Decode -
func (l *unionDecoder) Decode(data gjson.Result, path string, nm *meta.NodeMetadata, metadata meta.Metadata, isRoot bool) (*Node, error) {
	node := Node{
		Type:     nm.Type,
		Prim:     nm.Prim,
		Children: make([]*Node, 0),
	}

	if data.Value() == nil {
		return &node, nil
	}
	for i, arg := range nm.Args {
		argPath := strings.TrimPrefix(arg, path+"/")
		unionPath, err := getGJSONPathUnion(argPath, data)
		if err != nil {
			continue
		}
		argJSON := data.Get(unionPath)
		if argJSON.Exists() {
			argNode, err := l.parent.Convert(argJSON, arg, metadata, false)
			if err != nil {
				return nil, err
			}
			argMeta := metadata[arg]
			argNode.Name = argMeta.GetName(i)
			node.Children = append(node.Children, argNode)
			return &node, nil
		}
	}
	node.Name = metadata[path].GetName(-1)
	return &node, nil
}
