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
		if strings.Contains(a, "_slash_") || strings.Contains(a, ":_entries") || strings.Contains(a, `@\w+_slash_1`) {
			return true
		}
	}

	return false
}

func (l liquidity) CheckEntries(entry string) bool {
	return strings.Contains(entry, "_Liq_entry")
}
