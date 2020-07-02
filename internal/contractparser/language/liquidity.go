package language

import (
	"strings"

	"github.com/baking-bad/bcdhub/internal/contractparser/node"
	"github.com/tidwall/gjson"
)

type liquidity struct{}

func (l liquidity) DetectInCode(n node.Node) bool {
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

func (l liquidity) DetectInParameter(n node.Node) bool {
	if !n.HasAnnots() {
		return false
	}

	for _, entrypoint := range n.Annotations {
		if entrypoint[0] != '%' {
			continue
		}

		if strings.Contains(entrypoint, "_Liq_entry") {
			return true
		}
	}

	return false
}

func (l liquidity) DetectInFirstPrim(val gjson.Result) bool {
	return false
}
