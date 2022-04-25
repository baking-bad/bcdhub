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

// Minter -
type Minter struct {
	script *ast.Script
}

// NewMinter -
func NewMinter(ctx context.Context, rpc noderpc.INode) (*Minter, error) {
	contract := new(Minter)
	script, err := rpc.GetScriptJSON(ctx, contract.Address(), 0)
	if err != nil {
		return nil, err
	}
	contract.script = script.Code

	return contract, nil
}

// Address -
func (minter *Minter) Address() string {
	return "KT1S95Dyj2QrJpSnAbHRUSUZr7DhuFqssrog"
}

// HasHandler -
func (minter *Minter) HasHandler(entrypoint string) bool {
	_, ok := map[string]struct{}{
		mint: {},
	}[entrypoint]
	return ok
}

// Handler -
func (minter *Minter) Handler(parameters *types.Parameters) ([]tokenbalance.TokenBalance, error) {
	parameter, err := minter.script.ParameterType()
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
			if ownerPair, ok := node.(*ast.Pair); ok {
				if tokenPair, ok := ownerPair.Args[0].(*ast.Pair); ok {
					if tokenID, ok := tokenPair.Args[0].(*ast.Nat); ok {
						if owner, ok := ownerPair.Args[0].(*ast.Address); ok {
							address := forge.DecodeString(owner.GetValue().(string))
							if address == "" {
								return nil, nil
							}
							tID := tokenID.GetValue().(*types.BigInt)
							return []tokenbalance.TokenBalance{
								{
									TokenID: tID.Uint64(),
									Address: address,
									Value:   decimal.NewFromInt(1),
								},
							}, nil
						}

					}
				}
			}
		default:
		}
	}

	return nil, nil
}
