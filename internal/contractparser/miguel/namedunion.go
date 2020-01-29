package miguel

import (
	"strings"

	"github.com/aopoltorzhicky/bcdhub/internal/contractparser/meta"
	"github.com/tidwall/gjson"
)

type namedUnionDecoder struct{}

// Decode -
func (l *namedUnionDecoder) Decode(node gjson.Result, path string, nm *meta.NodeMetadata, metadata meta.Metadata) (interface{}, error) {
	res := make(map[string]interface{})
	for _, arg := range nm.Args {
		argPath := strings.TrimPrefix(arg, path+"/")
		unionPath := getGJSONPathUnion(argPath, node)
		argNode := node.Get(unionPath)
		name := meta.GetName(metadata[arg])

		if argNode.Exists() {
			data, err := michelineNodeToMiguel(argNode, arg, metadata)
			if err != nil {
				return nil, err
			}
			res[name] = data
			return res, nil
		}
	}

	name := meta.GetName(metadata[path])
	res[name] = nil
	return res, nil
}
