package newmiguel

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/baking-bad/bcdhub/internal/contractparser/consts"
	"github.com/baking-bad/bcdhub/internal/contractparser/meta"
	"github.com/pkg/errors"
	"github.com/tidwall/gjson"
)

type enumDecoder struct{}

// Decode -
func (l *enumDecoder) Decode(data gjson.Result, path string, nm *meta.NodeMetadata, metadata meta.Metadata, isRoot bool) (*Node, error) {
	node := Node{
		Type:     nm.Type,
		Prim:     nm.Prim,
		Children: make([]*Node, 0),
	}

	var tail string
	end := false
	for !end {
		prim := data.Get("prim|@lower").String()
		data = data.Get("args.0")
		switch prim {
		case consts.LEFT:
			tail += "/0"
		case consts.RIGHT:
			tail += "/1"
		default:
			end = true
		}
	}

	valNode, ok := metadata[path+tail]
	if !ok {
		return nil, errors.Errorf("Unknown enum path: %s", path+tail)
	}

	switch {
	case valNode.Name != "":
		node.Value = valNode.Name
	case tail == "":
		node.Value = consts.UNIT
	default:
		bin := strings.Replace(tail, "/", "", -1)
		i, err := strconv.ParseInt(bin, 2, 64)
		if err != nil {
			return nil, err
		}
		node.Value = fmt.Sprintf("value_%d", i)
	}

	return &node, nil
}
