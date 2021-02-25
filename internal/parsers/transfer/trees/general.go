package trees

import "github.com/baking-bad/bcdhub/internal/bcd/ast"

// GetFA1_2Transfer -
func GetFA1_2Transfer() *ast.TypedAst {
	return getTypedTree(fa1_2Transfer)
}

// GetFA2Transfer -
func GetFA2Transfer() *ast.TypedAst {
	return getTypedTree(fa2Transfer)
}

func getTypedTree(node ast.Node) *ast.TypedAst {
	return &ast.TypedAst{
		Nodes: []ast.Node{
			ast.Copy(node),
		},
	}
}
