package miguel

import (
	"strings"

	"github.com/baking-bad/bcdhub/internal/contractparser/meta"
	"github.com/tidwall/gjson"
)

type namedTupleDecoder struct{}

// Decode -
func (l *namedTupleDecoder) Decode(node gjson.Result, path string, nm *meta.NodeMetadata, metadata meta.Metadata, isRoot bool) (interface{}, error) {
	res := make(map[string]interface{})
	for i, arg := range nm.Args {
		argPath := strings.TrimPrefix(arg, path+"/")
		gjsonPath := GetGJSONPath(argPath)
		argNode := node.Get(gjsonPath)

		name := metadata.GetFieldName(arg, i)

		if argNode.Exists() {
			data, err := michelineNodeToMiguel(argNode, arg, metadata, false)
			if err != nil {
				return nil, err
			}
			res[name] = data
		} else {
			res[name] = nil
		}
	}
	return res, nil
}
