package contracts

import (
	"github.com/baking-bad/bcdhub/internal/bcd/ast"
	"github.com/baking-bad/bcdhub/internal/bcd/forge"
	"github.com/baking-bad/bcdhub/internal/bcd/types"
	"github.com/baking-bad/bcdhub/internal/noderpc"
	"github.com/baking-bad/bcdhub/internal/parsers/tokenbalance"
	"github.com/shopspring/decimal"
)

// KUSD -
type KUSD struct {
	script *ast.Script
}

// NewKUSD -
func NewKUSD(rpc noderpc.INode) (*KUSD, error) {
	contract := new(KUSD)

	script, err := rpc.GetScriptJSON(contract.Address(), 0)
	if err != nil {
		return nil, err
	}
	contract.script = script.Code

	return contract, nil
}

// Address -
func (kusd *KUSD) Address() string {
	return "KT1K9gCRgaLRFKTErYt1wVxA3Frb9FjasjTV"
}

// HasHandler -
func (kusd *KUSD) HasHandler(entrypoint string) bool {
	_, ok := map[string]struct{}{
		mint: {},
		burn: {},
	}[entrypoint]
	return ok
}

// Handler -
func (kusd *KUSD) Handler(parameters *types.Parameters) ([]tokenbalance.TokenBalance, error) {
	parameter, err := kusd.script.ParameterType()
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
					}
				}
			}
		}
	}

	return nil, nil
}
