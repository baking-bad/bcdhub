package trees

import (
	"github.com/baking-bad/bcdhub/internal/bcd"
	"github.com/baking-bad/bcdhub/internal/bcd/ast"
	"github.com/baking-bad/bcdhub/internal/bcd/forge"
	"github.com/baking-bad/bcdhub/internal/bcd/types"
	"github.com/baking-bad/bcdhub/internal/models/operation"
	"github.com/baking-bad/bcdhub/internal/models/transfer"
)

// MakeFa1_2Transfers -
func MakeFa1_2Transfers(tree ast.Node, operation operation.Operation) ([]*transfer.Transfer, error) {
	if tree == nil {
		return nil, nil
	}
	var err error

	t := transfer.EmptyTransfer(operation)
	pair := tree.(*ast.Pair)
	from := pair.Args[0].GetValue().(string)
	t.From, err = getAddress(from)
	if err != nil {
		return nil, err
	}
	toPair := pair.Args[1].(*ast.Pair)
	to := toPair.Args[0].GetValue().(string)
	t.To, err = getAddress(to)
	if err != nil {
		return nil, err
	}
	i := toPair.Args[1].GetValue().(*types.BigInt)
	t.AmountBigInt.Set(i.Int)
	return []*transfer.Transfer{t}, nil
}

func getAddress(address string) (string, error) {
	if bcd.IsAddress(address) {
		return address, nil
	}
	return forge.UnforgeAddress(address)
}
