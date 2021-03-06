package trees

import (
	"github.com/baking-bad/bcdhub/internal/bcd/ast"
	"github.com/baking-bad/bcdhub/internal/bcd/types"
	"github.com/baking-bad/bcdhub/internal/models/operation"
	"github.com/baking-bad/bcdhub/internal/models/transfer"
)

// MakeFa2Transfers -
func MakeFa2Transfers(tree ast.Node, operation operation.Operation) ([]*transfer.Transfer, error) {
	if tree == nil {
		return nil, nil
	}
	transfers := make([]*transfer.Transfer, 0)
	list := tree.(*ast.List)
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
