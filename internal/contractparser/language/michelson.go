package language

import (
	"regexp"

	"github.com/baking-bad/bcdhub/internal/contractparser/node"
	"github.com/tidwall/gjson"
)

type michelson struct{}

var michelsonVarAnnotation = regexp.MustCompile(`^@[A-Za-z]{1}[A-Za-z_0-9]+$`)
var michelsonStorageAnnotation = regexp.MustCompile(`^:[A-Za-z]{1}[A-Za-z_0-9]+$`)

func (l michelson) DetectInCode(n node.Node) bool {
	if !n.HasAnnots() {
		return false
	}

	for _, a := range n.Annotations {
		if a == "@%%" || a == "%@" || a == "@%" {
			return true
		}

		if michelsonVarAnnotation.MatchString(a) {
			return true
		}
	}

	return false
}

func (l michelson) DetectInParameter(n node.Node) bool {
	return false
}

func (l michelson) DetectInFirstPrim(val gjson.Result) bool {
	return false
}

// DetectMichelsonInStorage - costil++
func DetectMichelsonInStorage(val gjson.Result) string {
	if root := val.Get("annots"); root.Exists() {
		if michelsonStorageAnnotation.MatchString(root.String()) {
			return LangMichelson
		}
	}
	return LangUnknown
}
