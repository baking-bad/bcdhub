package language

import (
	"strings"

	"github.com/baking-bad/bcdhub/internal/contractparser/node"
)

type scaml struct{}

func (l scaml) Tag() string {
	return LangSCaml
}

func (l scaml) DetectInCode(n node.Node) bool {
	str := n.GetString()
	return strings.Contains(str, "Option.get") || strings.Contains(str, "Sum.get-left")
}

func (l scaml) DetectInParameter(n node.Node) bool {
	return false
}
