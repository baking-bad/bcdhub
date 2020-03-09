package language

import (
	"strings"

	"github.com/aopoltorzhicky/bcdhub/internal/contractparser/meta"
	"github.com/aopoltorzhicky/bcdhub/internal/contractparser/node"
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
		if strings.Contains(e.Name, "_Liq_entry") {
			return LangLiquidity
		}
	}
	return LangUnknown
}
