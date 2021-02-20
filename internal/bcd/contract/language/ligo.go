package language

import (
	"strconv"
	"strings"

	"github.com/baking-bad/bcdhub/internal/bcd/base"
)

type ligo struct{}

func (l ligo) DetectInCode(node *base.Node) bool {
	if len(node.Annots) > 0 {
		for _, a := range node.Annots {
			if len(a) < 2 {
				continue
			}
			if a[0] == '%' && isDigit(a[1:]) {
				return true
			}
		}
	}
	if node.StringValue == nil {
		return false
	}
	return hasLIGOKeywords(*node.StringValue)
}

func (l ligo) DetectInParameter(node *base.Node) bool {
	return false
}

// DetectInFirstPrim -
func (l ligo) DetectInFirstPrim(node *base.Node) bool {
	return false
}

func isDigit(input string) bool {
	_, err := strconv.ParseUint(input, 10, 32)
	return err == nil
}

func hasLIGOKeywords(s string) bool {
	ligoKeywords := []string{
		"GET_FORCE",
		"get_force",
		"MAP FIND",
		"failed assertion",
	}

	for _, keyword := range ligoKeywords {
		if s == keyword {
			return true
		}
	}

	return strings.Contains(s, "get_entrypoint") || strings.Contains(s, "get_contract")
}
