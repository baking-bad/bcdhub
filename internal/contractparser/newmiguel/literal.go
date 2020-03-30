package newmiguel

import (
	"fmt"
	"time"

	"github.com/baking-bad/bcdhub/internal/contractparser/consts"
	"github.com/baking-bad/bcdhub/internal/contractparser/meta"
	"github.com/baking-bad/bcdhub/internal/contractparser/unpack"
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
func (l *literalDecoder) Decode(jsonData gjson.Result, path string, nm *meta.NodeMetadata, metadata meta.Metadata, isRoot bool) (*Node, error) {
	node := Node{
		Prim: nm.Prim,
		Type: nm.Type,
	}

	if jsonData.Value() == nil {
		return &node, nil
	}
	switch nm.Type {
	case consts.MUTEZ, consts.NAT, consts.STRING, consts.INT:
		data, err := l.simple.Decode(jsonData, path, nm, metadata, false)
		if err != nil {
			return nil, err
		}
		node.Value = data
	case consts.BYTES:
		node.Value = unpack.Bytes(jsonData.Get(consts.BYTES).String())
	case consts.ADDRESS:
		if jsonData.Get(consts.BYTES).Exists() {
			data, err := unpack.Address(jsonData.Get(consts.BYTES).String())
			if err != nil {
				return nil, err
			}
			node.Value = data
		} else {
			node.Value = jsonData.Get(consts.STRING).String()
		}
	case consts.CONTRACT:
		if jsonData.Get(consts.BYTES).Exists() {
			data, err := unpack.Contract(jsonData.Get(consts.BYTES).String())
			if err != nil {
				return nil, err
			}
			node.Value = data
		} else {
			node.Value = jsonData.Get(consts.STRING).String()
		}
	case consts.KEYHASH:
		if jsonData.Get(consts.BYTES).Exists() {
			data, err := unpack.KeyHash(jsonData.Get(consts.BYTES).String())
			if err != nil {
				return nil, err
			}
			node.Value = data
		} else {
			node.Value = jsonData.Get(consts.STRING).String()
		}
	case consts.KEY:
		if jsonData.Get(consts.BYTES).Exists() {
			data, err := unpack.PublicKey(jsonData.Get(consts.BYTES).String())
			if err != nil {
				return nil, err
			}
			node.Value = data
		} else {
			node.Value = jsonData.Get(consts.STRING).String()
		}
	case consts.SIGNATURE:
		if jsonData.Get(consts.BYTES).Exists() {
			data, err := unpack.Signature(jsonData.Get(consts.BYTES).String())
			if err != nil {
				return nil, err
			}
			node.Value = data
		} else {
			node.Value = jsonData.Get(consts.STRING).String()
		}
	case consts.CHAINID:
		if jsonData.Get(consts.BYTES).Exists() {
			data, err := unpack.ChainID(jsonData.Get(consts.BYTES).String())
			if err != nil {
				return nil, err
			}
			node.Value = data
		} else {
			node.Value = jsonData.Get(consts.STRING).String()
		}
	case consts.TIMESTAMP:
		if jsonData.Get(consts.INT).Exists() {
			intVal := jsonData.Get(consts.INT).Int()
			if 253402300799 > intVal { // 31 December 9999 23:59:59 Golang time restriction
				node.Value = time.Unix(intVal, 0).UTC().String()
			} else {
				node.Value = fmt.Sprintf("%d", intVal)
			}
		} else if jsonData.Get(consts.STRING).Exists() {
			node.Value = jsonData.Get(consts.STRING).Time().UTC().String()
		}
	case consts.BOOL:
		node.Value = jsonData.Get("prim").String() != "False"
	case consts.UNIT:
		node.Value = nil
	}
	return &node, nil
}
