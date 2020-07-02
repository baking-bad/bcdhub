package language

import (
	"strings"

	"github.com/baking-bad/bcdhub/internal/contractparser/node"
	"github.com/tidwall/gjson"
)

type scaml struct{}

func (l scaml) DetectInCode(n node.Node) bool {
	str := n.GetString()
	return strings.Contains(str, "Option.get") || strings.Contains(str, "Sum.get-left")
}

func (l scaml) DetectInParameter(n node.Node) bool {
	return false
}

func (l scaml) DetectInFirstPrim(val gjson.Result) bool {
	return val.IsArray() && len(val.Array()) == 0
}
