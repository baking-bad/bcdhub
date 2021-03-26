package tokenbalance

import (
	"math/big"

	"github.com/baking-bad/bcdhub/internal/bcd/ast"
	"github.com/baking-bad/bcdhub/internal/bcd/forge"
	"github.com/baking-bad/bcdhub/internal/bcd/types"
)

// MultiAsset -
type MultiAsset struct {
	ReturnType *ast.TypedAst
}

// NewMultiAssetBalance -
func NewMultiAssetBalance() MultiAsset {
	node, _ := ast.NewTypedAstFromString(`{ "prim": "map", "args": [ { "prim": "pair", "args": [{ "prim": "address"}, {"prim": "nat"}] }, { "prim" : "nat" } ] }`)
	return MultiAsset{
		ReturnType: node,
	}
}

// NewMultiAssetUpdate -
func NewMultiAssetUpdate() MultiAsset {
	node, _ := ast.NewTypedAstFromString(`{ "prim": "map", "args": [ { "prim": "pair", "args": [{ "prim": "address"}, {"prim": "nat"}] }, { "prim" : "int" } ] }`)
	return MultiAsset{
		ReturnType: node,
	}
}

// GetReturnType -
func (p MultiAsset) GetReturnType() *ast.TypedAst {
	return p.ReturnType
}

// Parse -
func (p MultiAsset) Parse(data []byte) ([]TokenBalance, error) {
	m, err := getMap(p.ReturnType, data)
	if err != nil {
		return nil, err
	}

	balances := make([]TokenBalance, 0)
	err = m.Data.Range(func(key, value ast.Comparable) (bool, error) {
		val := value.(ast.Node)
		pair := key.(*ast.Pair)

		balance := val.GetValue().(*types.BigInt)
		tokenID := pair.Args[1].GetValue().(*types.BigInt)
		address := forge.DecodeString(pair.Args[0].GetValue().(string))
		if address == "" {
			return false, nil
		}

		amount, _ := new(big.Float).SetInt(balance.Int).Float64()
		balances = append(balances, TokenBalance{
			Value:   amount,
			Address: address,
			TokenID: tokenID.Uint64(),
		})
		return false, nil
	})
	if err != nil {
		return nil, err
	}

	return balances, nil
}
