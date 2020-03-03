package miguel

import (
	"fmt"

	"github.com/aopoltorzhicky/bcdhub/internal/contractparser/consts"
	"github.com/aopoltorzhicky/bcdhub/internal/contractparser/meta"
	"github.com/tidwall/gjson"
)

type optionDecoder struct {
	simple simpleDecoder
}

func newOptionDecoder() *optionDecoder {
	return &optionDecoder{
		simple: simpleDecoder{},
	}
}

// Decode -
func (d *optionDecoder) Decode(node gjson.Result, path string, nm *meta.NodeMetadata, metadata meta.Metadata, isRoot bool) (interface{}, error) {
	prim := node.Get("prim|@lower").String()
	switch prim {
	case consts.NONE:
		return nil, nil
	case consts.SOME:
		arg := node.Get("args.0")
		return d.simple.Decode(arg, path+"/0", nm, metadata, false)
	default:
		return nil, fmt.Errorf("optionDecoder.Decode: Unknown prim value %s", prim)
	}
}
