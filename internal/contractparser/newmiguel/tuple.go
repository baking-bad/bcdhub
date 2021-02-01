package newmiguel

import (
	"strings"

	"github.com/baking-bad/bcdhub/internal/contractparser/meta"
	"github.com/tidwall/gjson"
)

type tupleDecoder struct {
	parent *miguel
}

// Decode -
func (l *tupleDecoder) Decode(data gjson.Result, path string, nm *meta.NodeMetadata, metadata meta.Metadata, isRoot bool) (*Node, error) {
	node := Node{
		Prim:     nm.Prim,
		Type:     nm.Type,
		Name:     &(nm.Name),
		Children: make([]*Node, 0),
	}
	if data.Value() == nil {
		return &node, nil
	}
	for _, arg := range nm.Args {
		argPath := strings.TrimPrefix(arg, path+"/")
		gjsonPath := GetGJSONPath(argPath)
		argJSON := data.Get(gjsonPath)
		if argJSON.Exists() {
			argNode, err := l.parent.Convert(argJSON, arg, metadata, false)
			if err != nil {
				return nil, err
			}
			node.Children = append(node.Children, argNode)
		} else {
			argMeta := metadata[arg]
			node.Children = append(node.Children, &Node{
				Prim:     argMeta.Prim,
				Type:     argMeta.Type,
				Name:     &(argMeta.Name),
				Children: make([]*Node, 0),
			})
		}
	}
	return &node, nil
}
