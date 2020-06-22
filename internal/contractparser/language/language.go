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
	LangMichelson = "michelson"
	LangSCaml     = "scaml"
	LangUnknown   = "unknown"
)

type language interface {
	Tag() string
	DetectInCode(node.Node) bool
	DetectInParameter(node.Node) bool
}

var languages = []language{
	smartpy{},
	liquidity{},
	ligo{},
	lorentz{},
	michelson{},
	scaml{},
}

// GetFromCode -
func GetFromCode(n node.Node) string {
	for _, detector := range languages {
		if detector.DetectInCode(n) {
			return detector.Tag()
		}
	}
	return LangUnknown
}

// GetFromParameter -
func GetFromParameter(n node.Node) string {
	for _, detector := range languages {
		if detector.DetectInParameter(n) {
			return detector.Tag()
		}
	}
	return LangUnknown
}
