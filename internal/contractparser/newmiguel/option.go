package newmiguel

import (
	"fmt"

	"github.com/baking-bad/bcdhub/internal/contractparser/consts"
	"github.com/baking-bad/bcdhub/internal/contractparser/meta"
	"github.com/tidwall/gjson"
)

type optionDecoder struct{}

// Decode -
func (d *optionDecoder) Decode(node gjson.Result, path string, nm *meta.NodeMetadata, metadata meta.Metadata, isRoot bool) (*Node, error) {
	prim := node.Get("prim|@lower").String()
	switch prim {
	case consts.NONE:
		return &Node{}, nil
	case consts.SOME:
		arg := node.Get("args.0")
		return michelineNodeToMiguel(arg, path+"/o", metadata, false)
	default:
		return nil, fmt.Errorf("optionDecoder.Decode: Unknown prim value %s", prim)
	}
}
