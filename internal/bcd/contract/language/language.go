package language

import (
	"github.com/baking-bad/bcdhub/internal/contractparser/node"
	"github.com/tidwall/gjson"
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

// Priorities -
func Priorities() []string {
	return []string{
		LangSmartPy,
		LangLiquidity,
		LangLigo,
		LangLorentz,
		LangMichelson,
		LangSCaml,
		LangUnknown,
	}
}

type language interface {
	DetectInCode(node.Node) bool
	DetectInParameter(node.Node) bool
	DetectInFirstPrim(gjson.Result) bool
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

// GetFromFirstPrim -
func GetFromFirstPrim(n gjson.Result) string {
	for lang, detector := range languages {
		if detector.DetectInFirstPrim(n) {
			return lang
		}
	}
	return LangUnknown
}
