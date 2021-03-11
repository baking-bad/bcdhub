package language

import (
	"github.com/baking-bad/bcdhub/internal/bcd/base"
)

// Languages
const (
	LangSmartPy   = "smartpy"
	LangLiquidity = "liquidity"
	LangLigo      = "ligo"
	LangLorentz   = "lorentz"
	LangMichelson = "michelson"
	LangSCaml     = "scaml"
	LangUnknown   = "unknown"
)

var priorities = map[string]int{
	LangSmartPy:   6,
	LangLiquidity: 5,
	LangLigo:      4,
	LangLorentz:   3,
	LangMichelson: 2,
	LangSCaml:     1,
	LangUnknown:   0,
}

// GetPriority -
func GetPriority(lang string) int {
	if p, ok := priorities[lang]; ok {
		return p
	}
	return -1
}

type language interface {
	DetectInCode(node *base.Node) bool
	DetectInParameter(node *base.Node) bool
	DetectInFirstPrim(node *base.Node) bool
}

var languages = map[string]language{
	LangSmartPy:   smartpy{},
	LangLiquidity: liquidity{},
	LangLigo:      ligo{},
	LangLorentz:   lorentz{},
	LangMichelson: michelson{},
	LangSCaml:     scaml{},
}

// GetFromCode -
func GetFromCode(node *base.Node) string {
	for lang, detector := range languages {
		if detector.DetectInCode(node) {
			return lang
		}
	}
	return LangUnknown
}

// GetFromParameter -
func GetFromParameter(node *base.Node) string {
	for lang, detector := range languages {
		if detector.DetectInParameter(node) {
			return lang
		}
	}
	return LangUnknown
}

// GetFromFirstPrim -
func GetFromFirstPrim(node *base.Node) string {
	for lang, detector := range languages {
		if detector.DetectInFirstPrim(node) {
			return lang
		}
	}
	return LangUnknown
}
