package newmiguel

import (
	"strings"

	"github.com/baking-bad/bcdhub/internal/contractparser/meta"
	"github.com/tidwall/gjson"
)

type orDecoder struct{}

// Decode -
func (l *orDecoder) Decode(data gjson.Result, path string, nm *meta.NodeMetadata, metadata meta.Metadata, isRoot bool) (*Node, error) {
	node := Node{
		Type:     nm.Type,
		Prim:     nm.Type,
		Children: make([]*Node, 0),
	}
	root := metadata["0"]
	for i, arg := range root.Args {
		if !strings.HasPrefix(arg, path) {
			continue
		}
		argPath := strings.TrimPrefix(arg, path+"/")
		unionPath, err := getGJSONPathUnion(argPath, data)
		if err != nil {
			continue
		}
		argJSON := data.Get(unionPath)
		if argJSON.Exists() {
			argNode, err := michelineNodeToMiguel(argJSON, arg, metadata, false)
			if err != nil {
				return nil, err
			}

			argNode.Name = metadata.GetFieldName(arg, i)
			node.Children = append(node.Children, argNode)
			return &node, nil
		}
	}

	return nil, nil
}
