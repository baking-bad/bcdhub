package language

import (
	"regexp"
	"strings"

	"github.com/baking-bad/bcdhub/internal/contractparser/node"
	"github.com/tidwall/gjson"
)

type lorentz struct{}

const lorentzPrefix = "%epw"

var lorentzCamelCase = regexp.MustCompile(`([A-Z][a-z0-9]+)((\d)|([A-Z0-9][a-z0-9]+))*([A-Z])?`)

func (l lorentz) DetectInCode(n node.Node) bool {
	str := n.GetString()

	if str == "" {
		return false
	}

	return strings.Contains(str, "UStore") || strings.Contains(str, "Lorentz") || strings.Contains(str, "lorentz")
}

func (l lorentz) DetectInParameter(n node.Node) bool {
	if !n.HasAnnots() {
		return false
	}

	for _, entrypoint := range n.Annotations {
		if entrypoint[0] != '%' || len(entrypoint) < len(lorentzPrefix) {
			continue
		}

		if strings.HasPrefix(entrypoint, lorentzPrefix) && lorentzCamelCase.MatchString(entrypoint[len(lorentzPrefix):]) {
			return true
		}
	}

	return false
}

// DetectInFirstPrim -
func (l lorentz) DetectInFirstPrim(val gjson.Result) bool {
	return val.Get("prim").String() == "CAST"
}
