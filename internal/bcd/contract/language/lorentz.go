package language

import (
	"regexp"
	"strings"

	"github.com/baking-bad/bcdhub/internal/bcd/base"
)

type lorentz struct{}

const lorentzPrefix = "%epw"

var lorentzCamelCase = regexp.MustCompile(`([A-Z][a-z0-9]+)((\d)|([A-Z0-9][a-z0-9]+))*([A-Z])?`)

func (l lorentz) DetectInCode(node *base.Node) bool {
	if node.StringValue == nil {
		return false
	}
	str := *node.StringValue

	return strings.Contains(str, "UStore") || strings.Contains(str, "Lorentz") || strings.Contains(str, "lorentz")
}

func (l lorentz) DetectInParameter(node *base.Node) bool {
	if len(node.Annots) == 0 {
		return false
	}

	for _, entrypoint := range node.Annots {
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
func (l lorentz) DetectInFirstPrim(node *base.Node) bool {
	return node.Prim == "CAST"
}
