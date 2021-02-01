package newmiguel

import (
	"fmt"

	"github.com/baking-bad/bcdhub/internal/contractparser/consts"
	"github.com/baking-bad/bcdhub/internal/contractparser/formatter"
	"github.com/baking-bad/bcdhub/internal/contractparser/meta"
	"github.com/pkg/errors"
	"github.com/tidwall/gjson"
)

type mapDecoder struct {
	parent *miguel
}

// Decode -
func (l *mapDecoder) Decode(data gjson.Result, path string, nm *meta.NodeMetadata, metadata meta.Metadata, isRoot bool) (node *Node, err error) {
	node = &Node{
		Prim: nm.Prim,
		Type: nm.Type,
	}
	if data.Get("int").Exists() {
		node.Value = data.Get("int").Int()
		return
	}

	if data.IsArray() && len(data.Array()) == 0 && path == "0/0" {
		if nm.Prim == consts.BIGMAP {
			node.Value = 0
		}
		return
	}

	node.Children = make([]*Node, 0)
	if data.Value() == nil {
		return
	}
	gjsonPath := GetGJSONPath("k")
	keyJSON := data.Get(gjsonPath)

	for i, k := range keyJSON.Array() {
		key, err := l.parent.Convert(k, path+"/k", metadata, false)
		if err != nil {
			return nil, err
		}
		if key != nil {
			gjsonPath := fmt.Sprintf("%d.args.1", i)
			valJSON := data.Get(gjsonPath)
			var argNode *Node
			if valJSON.Exists() {
				argNode, err = l.parent.Convert(valJSON, path+"/v", metadata, false)
				if err != nil {
					return nil, err
				}
			}

			if key.Value == nil && len(key.Children) > 0 {
				key.Value, err = formatter.MichelineToMichelson(k, true, formatter.DefLineSize)
				if err != nil {
					return nil, err
				}
			}
			s, err := l.getKey(key)
			if err != nil {
				return nil, err
			}
			argNode.Name = &s
			node.Children = append(node.Children, argNode)
		}
	}

	return
}

func (l *mapDecoder) getKey(key *Node) (s string, err error) {
	switch kv := key.Value.(type) {
	case string:
		s = kv
	case int, int64:
		s = fmt.Sprintf("%d", kv)
	case bool:
		s = fmt.Sprintf("%t", kv)
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
		err = errors.Errorf("Invalid map key type: %v %T", key, key)
	}
	return
}
