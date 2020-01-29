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
func (l *literalDecoder) Decode(node gjson.Result, path string, nm *meta.NodeMetadata, metadata meta.Metadata) (interface{}, error) {
	switch nm.Type {
	case consts.KEYHASH, consts.BYTES, consts.CONTRACT, consts.NAT, consts.MUTEZ, consts.ADDRESS, consts.STRING, consts.TIMESTAMP, consts.KEY, consts.INT, consts.SIGNATURE:
		data, err := l.simple.Decode(node, path, nm, metadata)
		if err != nil {
			return nil, err
		}
		return data, nil
	case consts.BOOL:
		return node.Get("prim").Bool(), nil
	}
	return nil, nil
}
