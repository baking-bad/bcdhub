package newmiguel

import (
	"strings"

	"github.com/baking-bad/bcdhub/internal/contractparser/meta"
	"github.com/tidwall/gjson"
)

type namedTupleDecoder struct {
	parent *miguel
}

// Decode -
func (l *namedTupleDecoder) Decode(data gjson.Result, path string, nm *meta.NodeMetadata, metadata meta.Metadata, isRoot bool) (*Node, error) {
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
		gjsonPath := GetGJSONPath(argPath)
		argJSON := data.Get(gjsonPath)
		name := metadata.GetFieldName(arg, i)

		if argJSON.Exists() {
			argNode, err := l.parent.Convert(argJSON, arg, metadata, false)
			if err != nil {
				return nil, err
			}
			argNode.Name = name
			node.Children = append(node.Children, argNode)
		} else {
			node.Children = append(node.Children, &Node{
				Name:  name,
				Value: nil,
			})
		}
	}
	return &node, nil
}
