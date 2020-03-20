package newmiguel

import (
	"strings"

	"github.com/baking-bad/bcdhub/internal/contractparser/meta"
	"github.com/tidwall/gjson"
)

type tupleDecoder struct{}

// Decode -
func (l *tupleDecoder) Decode(data gjson.Result, path string, nm *meta.NodeMetadata, metadata meta.Metadata, isRoot bool) (*Node, error) {
	node := Node{
		Prim:     nm.Prim,
		Type:     nm.Type,
		Name:     nm.Name,
		Children: make([]*Node, 0),
	}

	for _, arg := range nm.Args {
		argPath := strings.TrimPrefix(arg, path+"/")
		gjsonPath := GetGJSONPath(argPath)
		argJSON := data.Get(gjsonPath)
		if argJSON.Exists() {
			argNode, err := michelineNodeToMiguel(argJSON, arg, metadata, false)
			if err != nil {
				return nil, err
			}
			node.Children = append(node.Children, argNode)
		} else {
			node.Children = append(node.Children, nil)
		}
	}
	return &node, nil
}
