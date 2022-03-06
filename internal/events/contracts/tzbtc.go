package contracts

import (
	"github.com/baking-bad/bcdhub/internal/bcd/ast"
	"github.com/baking-bad/bcdhub/internal/bcd/forge"
	"github.com/baking-bad/bcdhub/internal/bcd/types"
	"github.com/baking-bad/bcdhub/internal/noderpc"
	"github.com/baking-bad/bcdhub/internal/parsers/tokenbalance"
	"github.com/shopspring/decimal"
)

// TzBTC -
type TzBTC struct {
	script *ast.Script
}

// NewTzBTC -
func NewTzBTC(rpc noderpc.INode) (*TzBTC, error) {
	contract := new(TzBTC)

	script, err := rpc.GetScriptJSON(contract.Address(), 0)
	if err != nil {
		return nil, err
	}
	contract.script = script.Code

	return contract, nil
}

// Address -
func (tzbtc *TzBTC) Address() string {
	return "KT1PWx2mnDueood7fEmfbBDKx1D9BAnnXitn"
}

// HasHandler -
func (tzbtc *TzBTC) HasHandler(entrypoint string) bool {
	_, ok := map[string]struct{}{
		mint: {},
		burn: {},
	}[entrypoint]
	return ok
}

// ParameterEvent -
func (tzbtc *TzBTC) Handler(parameters *types.Parameters) ([]tokenbalance.TokenBalance, error) {
	parameter, err := tzbtc.script.ParameterType()
	if err != nil {
		return nil, err
	}
	typedAST, err := parameter.FromParameters(parameters)
	if err != nil {
		return nil, err
	}

	if node, entrypoint := typedAST.UnwrapAndGetEntrypointName(); node != nil {
		switch entrypoint {
		case mint:
			if pair, ok := node.(*ast.Pair); ok {
				if address, ok := pair.Args[0].(*ast.Address); ok {
					if amount, ok := pair.Args[1].(*ast.Nat); ok {
						balance := amount.GetValue().(*types.BigInt)
						value := forge.DecodeString(address.GetValue().(string))
						if value == "" {
							return nil, nil
						}
						return []tokenbalance.TokenBalance{
							{
								Address: value,
								TokenID: 0,
								Value:   decimal.NewFromBigInt(balance.Int, 0),
							},
						}, nil
					}
				}
			}
		case burn:
			if amount, ok := node.(*ast.Nat); ok {
				balance := amount.GetValue().(*types.BigInt)
				return []tokenbalance.TokenBalance{
					{
						TokenID: 0,
						Value:   decimal.NewFromBigInt(balance.Int, 0),
					},
				}, nil
			}
		}
	}

	return nil, nil
}
