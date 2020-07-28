package newmiguel

import (
	"fmt"

	"github.com/baking-bad/bcdhub/internal/contractparser/consts"
	"github.com/baking-bad/bcdhub/internal/contractparser/meta"
	"github.com/tidwall/gjson"
)

type optionDecoder struct {
	parent *miguel
}

// Decode -
func (d *optionDecoder) Decode(data gjson.Result, path string, nm *meta.NodeMetadata, metadata meta.Metadata, isRoot bool) (*Node, error) {
	prim := data.Get("prim|@lower").String()
	switch prim {
	case consts.NONE:
		return &Node{
			IsOption: true,
		}, nil
	case consts.SOME:
		arg := data.Get("args.0")
		node, err := d.parent.Convert(arg, path+"/o", metadata, false)
		if err != nil {
			return nil, err
		}
		node.IsOption = true
		return node, nil
	default:
		return nil, fmt.Errorf("optionDecoder.Decode: Unknown prim value %s", prim)
	}
}
