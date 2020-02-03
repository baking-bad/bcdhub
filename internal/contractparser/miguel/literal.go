package miguel

import (
	"github.com/aopoltorzhicky/bcdhub/internal/contractparser/consts"
	"github.com/aopoltorzhicky/bcdhub/internal/contractparser/meta"
	"github.com/tidwall/gjson"
)

type literalDecoder struct {
	simple simpleDecoder
}

func newLiteralDecoder() *literalDecoder {
	return &literalDecoder{
		simple: simpleDecoder{},
	}
}

// Decode -
func (l *literalDecoder) Decode(node gjson.Result, path string, nm *meta.NodeMetadata, metadata meta.Metadata) (map[string]interface{}, error) {
	switch nm.Type {
	case consts.KEYHASH, consts.BYTES, consts.CONTRACT, consts.MUTEZ, consts.NAT, consts.ADDRESS, consts.STRING, consts.KEY, consts.INT, consts.SIGNATURE:
		data, err := l.simple.Decode(node, path, nm, metadata)
		if err != nil {
			return nil, err
		}
		return map[string]interface{}{
			"value": data,
			"type":  nm.Type,
		}, nil
	case consts.TIMESTAMP:
		var value interface{}
		if node.Get(consts.INT).Exists() {
			value = node.Get(consts.INT).Int()
		} else if node.Get(consts.INT).Exists() {
			value = node.Get(consts.STRING).String()
		}
		return map[string]interface{}{
			"value": value,
			"type":  nm.Type,
		}, nil
	case consts.BOOL:
		return map[string]interface{}{
			"value": node.Get("prim").Bool(),
			"type":  nm.Type,
		}, nil
	}
	return nil, nil
}
