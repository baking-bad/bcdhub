package language

import "github.com/baking-bad/bcdhub/internal/contractparser/node"

type michelson struct{}

func (l michelson) Tag() string {
	return LangMichelson
}

func (l michelson) DetectInCode(n node.Node) bool {
	return false
}

func (l michelson) DetectInParameter(n node.Node) bool {
	return false
}
