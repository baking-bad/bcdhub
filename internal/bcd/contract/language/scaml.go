package language

import (
	"strings"

	"github.com/baking-bad/bcdhub/internal/bcd/base"
	"github.com/baking-bad/bcdhub/internal/bcd/consts"
)

type scaml struct{}

func (l scaml) DetectInCode(node *base.Node) bool {
	if node.StringValue == nil {
		return false
	}

	str := *node.StringValue
	return strings.Contains(str, "Option.get") || strings.Contains(str, "Sum.get-left")
}

func (l scaml) DetectInParameter(node *base.Node) bool {
	return false
}

func (l scaml) DetectInFirstPrim(node *base.Node) bool {
	return node.Prim == consts.PrimArray && len(node.Args) == 0
}
