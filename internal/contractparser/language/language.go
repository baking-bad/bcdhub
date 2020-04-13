package language

import (
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
	DetectInCode(node.Node) bool
	DetectInParameter(node.Node) bool
}

var languages = map[string]language{
	LangSmartPy:   smartpy{},
	LangLiquidity: liquidity{},
	LangLigo:      ligo{},
	LangLorentz:   lorentz{},
}

// GetFromCode -
func GetFromCode(n node.Node) string {
	for lang, detector := range languages {
		if detector.DetectInCode(n) {
			return lang
		}
	}
	return LangUnknown
}

// GetFromParameter -
func GetFromParameter(n node.Node) string {
	for lang, detector := range languages {
		if detector.DetectInParameter(n) {
			return lang
		}
	}
	return LangUnknown
}
