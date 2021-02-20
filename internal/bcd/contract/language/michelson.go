package language

import (
	"regexp"

	"github.com/baking-bad/bcdhub/internal/bcd/base"
)

type michelson struct{}

var michelsonVarAnnotation = regexp.MustCompile(`^@[A-Za-z]{1}[A-Za-z_0-9]+$`)
var michelsonStorageAnnotation = regexp.MustCompile(`^:[A-Za-z]{1}[A-Za-z_0-9]+$`)

func (l michelson) DetectInCode(node *base.Node) bool {
	if len(node.Annots) == 0 {
		return false
	}

	for _, a := range node.Annots {
		if a == "@%%" || a == "%@" || a == "@%" {
			return true
		}

		if michelsonVarAnnotation.MatchString(a) {
			return true
		}
	}

	return false
}

func (l michelson) DetectInParameter(node *base.Node) bool {
	return false
}

func (l michelson) DetectInFirstPrim(node *base.Node) bool {
	return false
}

// DetectMichelsonInStorage - costil++
func DetectMichelsonInStorage(node *base.Node) string {
	if len(node.Annots) > 0 {
		for i := range node.Annots {
			if michelsonStorageAnnotation.MatchString(node.Annots[i]) {
				return LangMichelson
			}
		}
	}
	return LangUnknown
}
