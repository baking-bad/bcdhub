package language

import (
	"strings"

	"github.com/aopoltorzhicky/bcdhub/internal/contractparser/node"
)

type lorentz struct{}

func (l lorentz) Detect(n node.Node) bool {
	str := n.GetString()

	if str == "" {
		return false
	}

	return strings.Contains(str, "UStore")
}
