package newmiguel

import (
	"fmt"

	"github.com/baking-bad/bcdhub/internal/contractparser/consts"
	"github.com/baking-bad/bcdhub/internal/contractparser/meta"
	"github.com/tidwall/gjson"
)

type simpleDecoder struct{}

func (l *simpleDecoder) Decode(data gjson.Result, path string, nm *meta.NodeMetadata, metadata meta.Metadata, isRoot bool) (interface{}, error) {
	prim := data.Get("prim|@lower").String()

	if prim == consts.NONE {
		return nil, nil
	}
	for k, v := range data.Map() {
		switch k {
		case consts.STRING, consts.BYTES, consts.INT:
			return v.String(), nil
		default:
			return nil, fmt.Errorf("Unknown simple type: %s %v", k, data)
		}
	}
	return nil, nil
}
