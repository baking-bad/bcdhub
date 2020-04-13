package language

import (
	"strings"

	"github.com/baking-bad/bcdhub/internal/contractparser/node"
)

type smartpy struct{}

func (l smartpy) DetectInCode(n node.Node) bool {
	str := n.GetString()

	if str == "" {
		return false
	}

	return strings.Contains(str, "SmartPy") ||
		strings.Contains(str, "self.") ||
		strings.Contains(str, "sp.") ||
		strings.Contains(str, "WrongCondition") ||
		strings.Contains(str, `Get-item:\d+`)
}

func (l smartpy) DetectInParameter(n node.Node) bool {
	return false
}
