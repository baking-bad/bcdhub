package ast

import "github.com/baking-bad/bcdhub/internal/bcd/forge"

// Forge -
func Forge(node Base, optimized bool) (string, error) {
	baseAST, err := node.ToBaseNode(optimized)
	if err != nil {
		panic(err)
	}
	return forge.ToString(baseAST)
}
