package miguel

import (
	"fmt"

	"github.com/aopoltorzhicky/bcdhub/internal/contractparser/consts"
	"github.com/aopoltorzhicky/bcdhub/internal/contractparser/meta"
	"github.com/tidwall/gjson"
)

type mapDecoder struct{}

// Decode -
func (l *mapDecoder) Decode(node gjson.Result, path string, nm *meta.NodeMetadata, metadata meta.Metadata, isRoot bool) (interface{}, error) {
	if node.Get("int").Exists() {
		return map[string]interface{}{
			"miguel_type":  consts.BIGMAP,
			"miguel_value": node.Get("int").Int(),
		}, nil
	}

	if node.IsArray() && len(node.Array()) == 0 && path == "0/0" {
		return map[string]interface{}{
			"miguel_type":  consts.BIGMAP,
			"miguel_value": 0,
		}, nil
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

			s, err := l.getKey(key)
			if err != nil {
				return nil, err
			}
			res[s] = value
		}
	}

	return res, nil
}

func (l *mapDecoder) getKey(key interface{}) (s string, err error) {
	switch kv := key.(type) {
	case string:
		s = kv
	case int, int64:
		s = fmt.Sprintf("%d", kv)
	case map[string]interface{}:
		s = fmt.Sprintf("%v", kv["miguel_value"])
	case []interface{}:
		s = ""
		for i, item := range kv {
			val := item.(map[string]interface{})
			if i != 0 {
				s += "@"
			}
			s += fmt.Sprintf("%v", val["miguel_value"])
		}
	default:
		err = fmt.Errorf("Invalid map key type: %v %T", key, key)
	}
	return
}
