package newmiguel

import (
	"github.com/baking-bad/bcdhub/internal/contractparser/consts"
	"github.com/baking-bad/bcdhub/internal/contractparser/meta"
	"github.com/tidwall/gjson"
)

type listDecoder struct{}

// Decode -
func (l *listDecoder) Decode(data gjson.Result, path string, nm *meta.NodeMetadata, metadata meta.Metadata, isRoot bool) (*Node, error) {
	node := Node{
		Prim:     nm.Prim,
		Type:     nm.Type,
		Name:     nm.Name,
		Children: make([]*Node, 0),
	}

	arr := data.Array()
	if len(arr) > 0 {
		subPath := "/l"
		if nm.Type == consts.SET {
			subPath = "/s"
		}
		for _, arg := range arr {
			argNode, err := michelineNodeToMiguel(arg, path+subPath, metadata, false)
			if err != nil {
				return nil, err
			}
			node.Children = append(node.Children, argNode)
		}
	}
	return &node, nil
}
