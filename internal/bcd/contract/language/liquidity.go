package language

import (
	"strings"

	"github.com/baking-bad/bcdhub/internal/bcd/base"
)

type liquidity struct{}

func (l liquidity) DetectInCode(node *base.Node) bool {
	if len(node.Annots) == 0 {
		return false
	}

	for _, a := range node.Annots {
		if strings.Contains(a, "_slash_") || strings.Contains(a, ":_entries") || strings.Contains(a, `@\w+_slash_1`) {
			return true
		}
	}

	return false
}

func (l liquidity) DetectInParameter(node *base.Node) bool {
	if len(node.Annots) == 0 {
		return false
	}

	for _, entrypoint := range node.Annots {
		if entrypoint[0] != '%' {
			continue
		}

		if strings.Contains(entrypoint, "_Liq_entry") {
			return true
		}
	}

	return false
}

func (l liquidity) DetectInFirstPrim(node *base.Node) bool {
	return false
}
