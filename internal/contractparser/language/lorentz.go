package language

import (
	"regexp"
	"strings"

	"github.com/baking-bad/bcdhub/internal/contractparser/node"
	"github.com/tidwall/gjson"
)

type lorentz struct{}

func (l lorentz) Tag() string {
	return LangLorentz
}

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

// DetectLorentzCast - checks that first prim in code is CAST
func DetectLorentzCast(val gjson.Result) string {
	if val.Get("0.0.prim").String() == "CAST" {
		return LangLorentz
	}

	return LangUnknown
}
