package miguel

import (
	"strings"

	"github.com/aopoltorzhicky/bcdhub/internal/contractparser/meta"
	"github.com/tidwall/gjson"
)

type orDecoder struct{}

// Decode -
func (l *orDecoder) Decode(node gjson.Result, path string, nm *meta.NodeMetadata, metadata meta.Metadata, isRoot bool) (interface{}, error) {
	res := make(map[string]interface{})

	root := metadata["0"]
	for i, arg := range root.Args {
		if !strings.HasPrefix(arg, path) {
			continue
		}
		argPath := strings.TrimPrefix(arg, path+"/")
		unionPath, err := getGJSONPathUnion(argPath, node)
		if err != nil {
			continue
		}
		argNode := node.Get(unionPath)
		if argNode.Exists() {
			data, err := michelineNodeToMiguel(argNode, arg, metadata, false)
			if err != nil {
				return nil, err
			}

			name := metadata.GetFieldName(arg, i)
			res[name] = data
			return res, nil
		}
	}

	return nil, nil
}
