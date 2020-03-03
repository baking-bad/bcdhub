package miguel

import (
	"strings"

	"github.com/aopoltorzhicky/bcdhub/internal/contractparser/meta"
	"github.com/tidwall/gjson"
)

type tupleDecoder struct{}

// Decode -
func (l *tupleDecoder) Decode(node gjson.Result, path string, nm *meta.NodeMetadata, metadata meta.Metadata, isRoot bool) (interface{}, error) {
	res := make([]interface{}, 0)
	for _, arg := range nm.Args {
		argPath := strings.TrimPrefix(arg, path+"/")
		gjsonPath := GetGJSONPath(argPath)
		argNode := node.Get(gjsonPath)
		if argNode.Exists() {
			data, err := michelineNodeToMiguel(argNode, arg, metadata, false)
			if err != nil {
				return nil, err
			}
			res = append(res, data)
		} else {
			res = append(res, nil)
		}
	}
	return res, nil
}
