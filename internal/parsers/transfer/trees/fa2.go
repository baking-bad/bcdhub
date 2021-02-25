package trees

import (
	"github.com/baking-bad/bcdhub/internal/bcd/ast"
	"github.com/baking-bad/bcdhub/internal/bcd/consts"
	"github.com/baking-bad/bcdhub/internal/bcd/types"
	"github.com/baking-bad/bcdhub/internal/models/operation"
	"github.com/baking-bad/bcdhub/internal/models/transfer"
)

var fa2Transfer = &ast.List{
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

// MakeFa2Transfers -
func MakeFa2Transfers(tree *ast.TypedAst, operation operation.Operation) ([]*transfer.Transfer, error) {
	if tree == nil || !tree.IsSettled() {
		return nil, nil
	}
	transfers := make([]*transfer.Transfer, 0)
	list := tree.Nodes[0].(*ast.List)
	for i := range list.Data {
		pair := list.Data[i].(*ast.Pair)
		from := pair.Args[0].GetValue().(string)
		toList := pair.Args[1].(*ast.List)
		for j := range toList.Data {
			var err error
			t := transfer.EmptyTransfer(operation)
			t.From, err = getAddress(from)
			if err != nil {
				return nil, err
			}
			toPair := toList.Data[j].(*ast.Pair)
			to := toPair.Args[0].GetValue().(string)
			t.To, err = getAddress(to)
			if err != nil {
				return nil, err
			}
			tokenPair := toPair.Args[1].(*ast.Pair)
			t.TokenID = tokenPair.Args[0].GetValue().(*types.BigInt).Int64()
			i := tokenPair.Args[1].GetValue().(*types.BigInt)
			t.AmountBigInt.Set(i.Int)
			transfers = append(transfers, t)
		}
	}
	return transfers, nil
}
