package trees

import (
	"github.com/baking-bad/bcdhub/internal/bcd"
	"github.com/baking-bad/bcdhub/internal/bcd/ast"
	"github.com/baking-bad/bcdhub/internal/bcd/forge"
	"github.com/baking-bad/bcdhub/internal/bcd/types"
	"github.com/baking-bad/bcdhub/internal/models/account"
	"github.com/baking-bad/bcdhub/internal/models/operation"
	"github.com/baking-bad/bcdhub/internal/models/transfer"
	modelTypes "github.com/baking-bad/bcdhub/internal/models/types"
	"github.com/shopspring/decimal"
)

// MakeFa1_2Transfers -
func MakeFa1_2Transfers(tree ast.Node, operation operation.Operation) ([]*transfer.Transfer, error) {
	if tree == nil {
		return nil, nil
	}
	var err error

	t := operation.EmptyTransfer()
	pair := tree.(*ast.Pair)
	from := pair.Args[0].GetValue().(string)
	fromAddr, err := getAddress(from)
	if err != nil {
		return nil, err
	}
	t.From = account.Account{
		Address: fromAddr,
		Type:    modelTypes.NewAccountType(fromAddr),
	}
	toPair := pair.Args[1].(*ast.Pair)
	to := toPair.Args[0].GetValue().(string)
	toAddr, err := getAddress(to)
	if err != nil {
		return nil, err
	}
	t.To = account.Account{
		Address: toAddr,
		Type:    modelTypes.NewAccountType(toAddr),
	}
	i := toPair.Args[1].GetValue().(*types.BigInt)
	t.Amount = decimal.NewFromBigInt(i.Int, 0)
	return []*transfer.Transfer{t}, nil
}

func getAddress(address string) (string, error) {
	if bcd.IsAddressLazy(address) {
		return address, nil
	}
	return forge.UnforgeAddress(address)
}
