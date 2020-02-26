package miguel

import (
	"strings"

	"github.com/aopoltorzhicky/bcdhub/internal/contractparser/meta"
	"github.com/tidwall/gjson"
)

type orDecoder struct{}

// Decode -
func (l *orDecoder) Decode(node gjson.Result, path string, nm *meta.NodeMetadata, metadata meta.Metadata) (interface{}, error) {
	res := make(map[string]interface{})

	root := metadata["0"]
	for _, arg := range root.Args {
		if !strings.HasPrefix(arg, path) {
			continue
		}
		argPath := strings.TrimPrefix(arg, path)
		unionPath := getGJSONPathUnion(argPath, node)
		argNode := node.Get(unionPath)
		if argNode.Exists() {
			data, err := michelineNodeToMiguel(argNode, arg, metadata)
			if err != nil {
				return nil, err
			}
			name := metadata[path].GetName()
			res[name] = data
			return res, nil
		}
	}

	return nil, nil
}
