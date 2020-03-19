package language

import (
	"regexp"
	"strings"

	"github.com/baking-bad/bcdhub/internal/contractparser/node"
)

type lorentz struct{}

var lorentzCamelCase = regexp.MustCompile(`([A-Z][a-z0-9]+)((\d)|([A-Z0-9][a-z0-9]+))*([A-Z])?`)

func (l lorentz) Detect(n node.Node) bool {
	str := n.GetString()

	if str == "" {
		return false
	}

	return strings.Contains(str, "UStore")
}

func (l lorentz) CheckEntries(entry string) bool {
	if !strings.HasPrefix(entry, "epw") {
		return false
	}

	return lorentzCamelCase.MatchString(entry[3:])
}
