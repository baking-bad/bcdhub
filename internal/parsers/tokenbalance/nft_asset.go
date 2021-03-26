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
	m, err := getMap(p.ReturnType, data)
	if err != nil {
		return nil, err
	}

	balances := make([]TokenBalance, 0)
	err = m.Data.Range(func(key, value ast.Comparable) (bool, error) {
		k := key.(*ast.Nat)
		tokenID := k.GetValue().(*types.BigInt)
		balance := big.NewInt(0)

		var address string
		switch t := value.(type) {
		case *ast.Address:
			if s, ok := t.Value.(string); ok {
				address = forge.DecodeString(s)
				balance.SetInt64(1)
			} else {
				return false, nil
			}
		case *ast.Option:
			if t.IsSome() {
				val := t.Type.(*ast.Address)
				if s, ok := val.Value.(string); ok {
					address = forge.DecodeString(s)
					balance.SetInt64(1)
				} else {
					return false, nil
				}
			}
		default:
			return false, nil
		}

		amount, _ := new(big.Float).SetInt(balance).Float64()
		balances = append(balances, TokenBalance{
			Value:          amount,
			Address:        address,
			TokenID:        tokenID.Uint64(),
			IsExclusiveNFT: true,
		})
		return false, nil
	})
	if err != nil {
		return nil, err
	}

	return balances, nil
}
