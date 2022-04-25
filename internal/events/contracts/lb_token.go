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

// TzBTC -
type LBToken struct {
	script *ast.Script
}

// NewLBToken -
func NewLBToken(ctx context.Context, rpc noderpc.INode) (*LBToken, error) {
	contract := new(LBToken)
	script, err := rpc.GetScriptJSON(ctx, contract.Address(), 0)
	if err != nil {
		return nil, err
	}
	contract.script = script.Code
	return contract, nil
}

// Address -
func (lb *LBToken) Address() string {
	return "KT1AafHA1C1vk959wvHWBispY9Y2f3fxBUUo"
}

// HasHandler -
func (lb *LBToken) HasHandler(entrypoint string) bool {
	_, ok := map[string]struct{}{
		mintOrBurn: {},
	}[entrypoint]
	return ok
}

// ParameterEvent -
func (lb *LBToken) Handler(parameters *types.Parameters) ([]tokenbalance.TokenBalance, error) {
	parameter, err := lb.script.ParameterType()
	if err != nil {
		return nil, err
	}
	typedAST, err := parameter.FromParameters(parameters)
	if err != nil {
		return nil, err
	}

	if node, entrypoint := typedAST.UnwrapAndGetEntrypointName(); node != nil && entrypoint == mintOrBurn {
		if pair, ok := node.(*ast.Pair); ok {
			if address, ok := pair.Args[1].(*ast.Address); ok {
				if amount, ok := pair.Args[0].(*ast.Int); ok {
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
	}

	return nil, nil
}
