package contracts

import (
	"context"

	"github.com/baking-bad/bcdhub/internal/bcd/ast"
	"github.com/baking-bad/bcdhub/internal/bcd/forge"
	"github.com/baking-bad/bcdhub/internal/bcd/types"
	"github.com/baking-bad/bcdhub/internal/noderpc"
	"github.com/baking-bad/bcdhub/internal/parsers/tokenbalance"
	"github.com/shopspring/decimal"
)

// USDtz -
type USDtz struct {
	script *ast.Script
}

// NewUSDtz -
func NewUSDtz(ctx context.Context, rpc noderpc.INode) (*USDtz, error) {
	contract := new(USDtz)
	script, err := rpc.GetScriptJSON(ctx, contract.Address(), 0)
	if err != nil {
		return nil, err
	}
	contract.script = script.Code
	return contract, nil
}

// Address -
func (usdtz *USDtz) Address() string {
	return "KT1LN4LPSqTMS7Sd2CJw4bbDGRkMv2t68Fy9"
}

// HasHandler -
func (usdtz *USDtz) HasHandler(entrypoint string) bool {
	_, ok := map[string]struct{}{
		mint: {},
		burn: {},
	}[entrypoint]
	return ok
}

// Handler -
func (usdtz *USDtz) Handler(parameters *types.Parameters) ([]tokenbalance.TokenBalance, error) {
	parameter, err := usdtz.script.ParameterType()
	if err != nil {
		return nil, err
	}
	typedAST, err := parameter.FromParameters(parameters)
	if err != nil {
		return nil, err
	}

	if node, entrypoint := typedAST.UnwrapAndGetEntrypointName(); node != nil {
		if pair, ok := node.(*ast.Pair); ok {
			if address, ok := pair.Args[0].(*ast.Address); ok {
				if amount, ok := pair.Args[1].(*ast.Nat); ok {
					value := forge.DecodeString(address.GetValue().(string))
					if value == "" {
						return nil, nil
					}
					balance := amount.GetValue().(*types.BigInt)

					tb := tokenbalance.TokenBalance{
						TokenID: 0,
						Address: value,
						Value:   decimal.NewFromBigInt(balance.Int, 0),
					}
					switch entrypoint {
					case mint:
						return []tokenbalance.TokenBalance{tb}, nil
					case burn:
						tb.Value = tb.Value.Neg()
						return []tokenbalance.TokenBalance{tb}, nil
					default:
						return nil, nil
					}
				}
			}
		}

	}

	return nil, nil
}
