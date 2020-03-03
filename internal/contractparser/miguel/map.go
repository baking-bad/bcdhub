package miguel

import (
	"fmt"

	"github.com/aopoltorzhicky/bcdhub/internal/contractparser/meta"
	"github.com/tidwall/gjson"
)

type mapDecoder struct{}

// Decode -
func (l *mapDecoder) Decode(node gjson.Result, path string, nm *meta.NodeMetadata, metadata meta.Metadata, isRoot bool) (interface{}, error) {
	if node.Get("int").Exists() {
		return map[string]interface{}{}, nil
	}
	res := make(map[string]interface{})
	gjsonPath := GetGJSONPath("k")
	keyNode := node.Get(gjsonPath)

	for i, k := range keyNode.Array() {
		key, err := michelineNodeToMiguel(k, path+"/k", metadata, false)
		if err != nil {
			return nil, err
		}
		if key != nil {
			gjsonPath := fmt.Sprintf("%d.args.1", i)
			valNode := node.Get(gjsonPath)
			var value interface{}
			if valNode.Exists() {
				value, err = michelineNodeToMiguel(valNode, path+"/v", metadata, false)
				if err != nil {
					return nil, err
				}
			}

			switch kv := key.(type) {
			case string:
				res[kv] = value
			case int, int64:
				s := fmt.Sprintf("%d", kv)
				res[s] = value
			case map[string]interface{}:
				s := fmt.Sprintf("%v", kv["value"])
				res[s] = value
			}
		}
	}
	return res, nil
}
