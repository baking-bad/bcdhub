package miguel

import (
	"github.com/aopoltorzhicky/bcdhub/internal/contractparser/consts"
	"github.com/aopoltorzhicky/bcdhub/internal/contractparser/meta"
	"github.com/aopoltorzhicky/bcdhub/internal/contractparser/unpack"
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
func (l *literalDecoder) Decode(node gjson.Result, path string, nm *meta.NodeMetadata, metadata meta.Metadata) (interface{}, error) {
	switch nm.Type {
	case consts.CONTRACT, consts.MUTEZ, consts.NAT, consts.STRING, consts.INT, consts.SIGNATURE:
		data, err := l.simple.Decode(node, path, nm, metadata)
		if err != nil {
			return nil, err
		}
		return map[string]interface{}{
			"value": data,
			"type":  nm.Type,
		}, nil
	case consts.BYTES:
		data := unpack.Bytes(node.Get(consts.BYTES).String())
		return map[string]interface{}{
			"value": data,
			"type":  nm.Type,
		}, nil
	case consts.ADDRESS:
		if node.Get(consts.BYTES).Exists() {
			data, err := unpack.Address(node.Get(consts.BYTES).String())
			if err != nil {
				return nil, err
			}
			return map[string]interface{}{
				"value": data,
				"type":  nm.Type,
			}, nil
		}
		return map[string]interface{}{
			"value": node.Get(consts.STRING).String(),
			"type":  nm.Type,
		}, nil
	case consts.KEYHASH:
		if node.Get(consts.BYTES).Exists() {
			data, err := unpack.KeyHash(node.Get(consts.BYTES).String())
			if err != nil {
				return nil, err
			}
			return map[string]interface{}{
				"value": data,
				"type":  nm.Type,
			}, nil
		}
		return map[string]interface{}{
			"value": node.Get(consts.STRING).String(),
			"type":  nm.Type,
		}, nil
	case consts.KEY:
		if node.Get(consts.BYTES).Exists() {
			data, err := unpack.PublicKey(node.Get(consts.BYTES).String())
			if err != nil {
				return nil, err
			}
			return map[string]interface{}{
				"value": data,
				"type":  nm.Type,
			}, nil
		}
		return map[string]interface{}{
			"value": node.Get(consts.STRING).String(),
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
