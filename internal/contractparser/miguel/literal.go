package miguel

import (
	"time"

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
func (l *literalDecoder) Decode(node gjson.Result, path string, nm *meta.NodeMetadata, metadata meta.Metadata, isRoot bool) (interface{}, error) {
	switch nm.Type {
	case consts.CONTRACT, consts.MUTEZ, consts.NAT, consts.STRING, consts.INT, consts.SIGNATURE:
		data, err := l.simple.Decode(node, path, nm, metadata, false)
		if err != nil {
			return nil, err
		}
		return map[string]interface{}{
			"miguel_value": data,
			"miguel_type":  nm.Type,
		}, nil
	case consts.BYTES:
		data := unpack.Bytes(node.Get(consts.BYTES).String())
		return map[string]interface{}{
			"miguel_value": data,
			"miguel_type":  nm.Type,
		}, nil
	case consts.ADDRESS:
		if node.Get(consts.BYTES).Exists() {
			data, err := unpack.Address(node.Get(consts.BYTES).String())
			if err != nil {
				return nil, err
			}
			return map[string]interface{}{
				"miguel_value": data,
				"miguel_type":  nm.Type,
			}, nil
		}
		return map[string]interface{}{
			"miguel_value": node.Get(consts.STRING).String(),
			"miguel_type":  nm.Type,
		}, nil
	case consts.KEYHASH:
		if node.Get(consts.BYTES).Exists() {
			data, err := unpack.KeyHash(node.Get(consts.BYTES).String())
			if err != nil {
				return nil, err
			}
			return map[string]interface{}{
				"miguel_value": data,
				"miguel_type":  nm.Type,
			}, nil
		}
		return map[string]interface{}{
			"miguel_value": node.Get(consts.STRING).String(),
			"miguel_type":  nm.Type,
		}, nil
	case consts.KEY:
		if node.Get(consts.BYTES).Exists() {
			data, err := unpack.PublicKey(node.Get(consts.BYTES).String())
			if err != nil {
				return nil, err
			}
			return map[string]interface{}{
				"miguel_value": data,
				"miguel_type":  nm.Type,
			}, nil
		}
		return map[string]interface{}{
			"miguel_value": node.Get(consts.STRING).String(),
			"miguel_type":  nm.Type,
		}, nil

	case consts.TIMESTAMP:
		var value interface{}
		if node.Get(consts.INT).Exists() {
			intVal := node.Get(consts.INT).Int()
			if 253402300799 > intVal { // 31 December 9999 23:59:59 Golang time restriction
				value = time.Unix(intVal, 0).UTC()
			} else {
				value = intVal
			}
		} else if node.Get(consts.STRING).Exists() {
			value = node.Get(consts.STRING).Time().UTC()
		}
		return map[string]interface{}{
			"miguel_value": value,
			"miguel_type":  nm.Type,
		}, nil
	case consts.BOOL:
		return map[string]interface{}{
			"miguel_value": node.Get("prim").String() != "False",
			"miguel_type":  nm.Type,
		}, nil
	case consts.UNIT:
		return map[string]interface{}{
			"miguel_value": nil,
			"miguel_type":  nm.Type,
		}, nil
	}
	return nil, nil
}
