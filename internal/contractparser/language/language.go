package language

import (
	"github.com/baking-bad/bcdhub/internal/contractparser/meta"
	"github.com/baking-bad/bcdhub/internal/contractparser/node"
)

// Languages
const (
	LangSmartPy   = "smartpy"
	LangLiquidity = "liquidity"
	LangLigo      = "ligo"
	LangLorentz   = "lorentz"
	LangUnknown   = "michelson"
)

type language interface {
	Detect(node.Node) bool
}

var languages = map[string]language{
	LangSmartPy:   smartpy{},
	LangLiquidity: liquidity{},
	LangLigo:      ligo{},
	LangLorentz:   lorentz{},
}

// Get -
func Get(n node.Node) string {
	for lang, detector := range languages {
		if detector.Detect(n) {
			return lang
		}
	}
	return LangUnknown
}

// DetectInEntries -
func DetectInEntries(entries []meta.Entrypoint) string {
	for _, e := range entries {
		if new(liquidity).CheckEntries(e.Name) {
			return LangLiquidity
		}

		if new(lorentz).CheckEntries(e.Name) {
			return LangLorentz
		}
	}
	return LangUnknown
}
