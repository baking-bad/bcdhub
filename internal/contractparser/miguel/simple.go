package miguel

import (
	"fmt"

	"github.com/aopoltorzhicky/bcdhub/internal/contractparser/consts"
	"github.com/aopoltorzhicky/bcdhub/internal/contractparser/meta"
	"github.com/tidwall/gjson"
)

type simpleDecoder struct{}

func (l *simpleDecoder) Decode(node gjson.Result, path string, nm *meta.NodeMetadata, metadata meta.Metadata, isRoot bool) (interface{}, error) {
	prim := node.Get("prim|@lower").String()
	if prim == consts.NONE {
		return nil, nil
	}
	for k, v := range node.Map() {
		switch k {
		case consts.STRING, consts.BYTES, consts.INT:
			return v.String(), nil
		default:
			return nil, fmt.Errorf("Unknown simple type: %s %v", k, node)
		}
	}
	return nil, nil
}
