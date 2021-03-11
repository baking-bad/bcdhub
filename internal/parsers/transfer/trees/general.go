package trees

import (
	"github.com/baking-bad/bcdhub/internal/bcd/ast"
	"github.com/baking-bad/bcdhub/internal/bcd/consts"
)

// GetFA1_2Transfer -
func GetFA1_2Transfer() ast.Node {
	return &ast.Pair{
		Default: ast.NewDefault(consts.PAIR, 2, 0),
		Args: []ast.Node{
			&ast.Address{
				Default: ast.NewDefault(consts.ADDRESS, 0, 1),
			},
			&ast.Pair{
				Default: ast.NewDefault(consts.PAIR, 2, 1),
				Args: []ast.Node{
					&ast.Address{
						Default: ast.NewDefault(consts.ADDRESS, 0, 2),
					},
					&ast.Nat{
						Default: ast.NewDefault(consts.NAT, 0, 2),
					},
				},
			},
		},
	}
}

// GetFA2Transfer -
func GetFA2Transfer() ast.Node {
	return &ast.List{
		Default: ast.NewDefault(consts.LIST, 0, 0),
		Type: &ast.Pair{
			Default: ast.NewDefault(consts.PAIR, 2, 1),
			Args: []ast.Node{
				&ast.Address{
					Default: ast.NewDefault(consts.ADDRESS, 0, 2),
				},
				&ast.List{
					Default: ast.NewDefault(consts.LIST, 0, 2),
					Type: &ast.Pair{
						Default: ast.NewDefault(consts.PAIR, 2, 3),
						Args: []ast.Node{
							&ast.Address{
								Default: ast.NewDefault(consts.ADDRESS, 0, 4),
							},
							&ast.Pair{
								Default: ast.NewDefault(consts.PAIR, 2, 4),
								Args: []ast.Node{
									&ast.Nat{
										Default: ast.NewDefault(consts.NAT, 0, 5),
									},
									&ast.Nat{
										Default: ast.NewDefault(consts.NAT, 0, 5),
									},
								},
							},
						},
					},
				},
			},
		},
	}
}
