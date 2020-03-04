package language

import (
	"strings"

	"github.com/aopoltorzhicky/bcdhub/internal/contractparser/node"
)

type smartpy struct{}

func (l smartpy) Detect(n node.Node) bool {
	str := n.GetString()

	if str == "" {
		return false
	}

	return strings.Contains(str, "SmartPy") || strings.Contains(str, "self.") || strings.Contains(str, "sp.") || strings.Contains(str, "WrongCondition")
}
