package tokenbalance

import (
	"math/big"

	"github.com/baking-bad/bcdhub/internal/bcd/ast"
	"github.com/baking-bad/bcdhub/internal/bcd/forge"
	"github.com/baking-bad/bcdhub/internal/bcd/types"
)

// NftSingleAsset -
type NftSingleAsset struct {
	ReturnType *ast.TypedAst
}

// NewNftSingleAsset -
func NewNftSingleAsset() NftSingleAsset {
	node, _ := ast.NewTypedAstFromString(`{"prim":"map","args":[{"prim":"address"},{"prim":"nat"}]}`)
	return NftSingleAsset{
		ReturnType: node,
	}
}

// NewNftSingleAssetOption -
func NewNftSingleAssetOption() NftSingleAsset {
	node, _ := ast.NewTypedAstFromString(`{"prim":"map","args":[{"prim":"address"},{"prim":"option","args":[{"prim":"nat"}]}]}`)
	return NftSingleAsset{
		ReturnType: node,
	}
}

// GetReturnType -
func (p NftSingleAsset) GetReturnType() *ast.TypedAst {
	return p.ReturnType
}

// Parse -
func (p NftSingleAsset) Parse(data []byte) ([]TokenBalance, error) {
	m, err := getMap(p.ReturnType, data)
	if err != nil {
		return nil, err
	}

	balances := make([]TokenBalance, 0)
	err = m.Data.Range(func(key, value ast.Comparable) (bool, error) {
		k := key.(*ast.Address)
		var address string
		if s, ok := k.Value.(string); ok {
			address = forge.DecodeString(s)
		} else {
			return false, nil
		}

		balance := big.NewInt(0)

		switch t := value.(type) {
		case *ast.Nat:
			if s, ok := t.GetValue().(*types.BigInt); ok {
				balance.Set(s.Int)
			} else {
				return false, nil
			}
		case *ast.Option:
			if t.IsSome() {
				val := t.Type.(*ast.Address)
				if s, ok := val.GetValue().(*types.BigInt); ok {
					balance.Set(s.Int)
				} else {
					return false, nil
				}
			}
		default:
			return false, nil
		}

		balances = append(balances, TokenBalance{
			Value:   balance,
			Address: address,
			IsNFT:   true,
		})
		return false, nil
	})
	if err != nil {
		return nil, err
	}

	return balances, nil
}
