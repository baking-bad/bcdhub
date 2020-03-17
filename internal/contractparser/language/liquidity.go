package language

import (
	"strings"

	"github.com/baking-bad/bcdhub/internal/contractparser/node"
)

type liquidity struct{}

func (l liquidity) Detect(n node.Node) bool {
	if !n.HasAnnots() {
		return false
	}

	for _, a := range n.Annotations {
		if strings.Contains(a, "_slash_") {
			return true
		}
	}

	return false
}
