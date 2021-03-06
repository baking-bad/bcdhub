package tokenbalance

import (
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
	var node ast.UntypedAST
	if err := json.Unmarshal(data, &node); err != nil {
		return nil, err
	}

	if err := p.ReturnType.Settle(node); err != nil {
		return nil, err
	}

	balances := make([]TokenBalance, 0)

	m := p.ReturnType.Nodes[0].(*ast.Map)
	err := m.Data.Range(func(key, value ast.Comparable) (bool, error) {
		val := value.(ast.Node)
		pair := key.(*ast.Pair)

		balance := val.GetValue().(*types.BigInt)
		tokenID := pair.Args[1].GetValue().(*types.BigInt)
		address := forge.DecodeString(pair.Args[0].GetValue().(string))
		if address == "" {
			return false, nil
		}

		balances = append(balances, TokenBalance{
			Value:   balance.Int,
			Address: address,
			TokenID: tokenID.Int64(),
		})
		return false, nil
	})
	if err != nil {
		return nil, err
	}

	return balances, nil
}
