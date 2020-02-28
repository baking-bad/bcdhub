package miguel

import (
	"strings"

	"github.com/aopoltorzhicky/bcdhub/internal/contractparser/meta"
	"github.com/tidwall/gjson"
)

type namedTupleDecoder struct{}

// Decode -
func (l *namedTupleDecoder) Decode(node gjson.Result, path string, nm *meta.NodeMetadata, metadata meta.Metadata) (interface{}, error) {
	res := make(map[string]interface{})
	for i, arg := range nm.Args {
		argPath := strings.TrimPrefix(arg, path+"/")
		gjsonPath := GetGJSONPath(argPath)
		argNode := node.Get(gjsonPath)
		name := metadata[arg].GetName(i)

		if argNode.Exists() {
			data, err := michelineNodeToMiguel(argNode, arg, metadata)
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
