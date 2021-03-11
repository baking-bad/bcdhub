package tokenbalance

import (
	"math/big"

	"github.com/baking-bad/bcdhub/internal/bcd/ast"
	"github.com/baking-bad/bcdhub/internal/bcd/forge"
	"github.com/baking-bad/bcdhub/internal/bcd/types"
)

// NftAsset -
type NftAsset struct {
	ReturnType *ast.TypedAst
}

// NewNftAsset -
func NewNftAsset() NftAsset {
	node, _ := ast.NewTypedAstFromString(`{"prim":"map","args":[{"prim":"nat"},{"prim":"address"}]}`)
	return NftAsset{
		ReturnType: node,
	}
}

// NewNftAssetOption -
func NewNftAssetOption() NftAsset {
	node, _ := ast.NewTypedAstFromString(`{"prim":"map","args":[{"prim":"nat"},{"prim":"option","args":[{"prim":"address"}]}]}`)
	return NftAsset{
		ReturnType: node,
	}
}

// GetReturnType -
func (p NftAsset) GetReturnType() *ast.TypedAst {
	return p.ReturnType
}

// Parse -
func (p NftAsset) Parse(data []byte) ([]TokenBalance, error) {
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
		k := key.(*ast.Nat)
		tokenID := k.GetValue().(*types.BigInt)
		balance := big.NewInt(0)

		var address string
		switch t := value.(type) {
		case *ast.Address:
			address = forge.DecodeString(t.Value.(string))
			balance.SetInt64(1)
		case *ast.Option:
			if t.IsSome() {
				val := t.Type.(*ast.Address)
				address = forge.DecodeString(val.Value.(string))
				balance.SetInt64(1)
			}
		default:
			return false, nil
		}

		balances = append(balances, TokenBalance{
			Value:   balance,
			Address: address,
			TokenID: tokenID.Int64(),
			IsNFT:   true,
		})
		return false, nil
	})
	if err != nil {
		return nil, err
	}

	return balances, nil
}
