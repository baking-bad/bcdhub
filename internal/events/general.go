package events

import (
	"context"
	stdJSON "encoding/json"
	"fmt"
	"strings"

	"github.com/baking-bad/bcdhub/internal/bcd/ast"
	"github.com/baking-bad/bcdhub/internal/logger"
	"github.com/baking-bad/bcdhub/internal/models/types"
	"github.com/baking-bad/bcdhub/internal/noderpc"
	"github.com/baking-bad/bcdhub/internal/parsers/tokenbalance"
)

// Event -
type Event interface {
	GetCode() ([]byte, error)
	Parse(response noderpc.RunCodeResponse) []tokenbalance.TokenBalance
	Normalize(parameter *ast.TypedAst) []byte
}

// Args -
type Args struct {
	Network                  types.Network
	Protocol                 string
	Parameters               *ast.TypedAst
	Source                   string
	Initiator                string
	Entrypoint               string
	ChainID                  string
	HardGasLimitPerOperation int64
	Amount                   int64
}

// Sections -
type Sections struct {
	Parameter  stdJSON.RawMessage
	ReturnType stdJSON.RawMessage
	Code       stdJSON.RawMessage
}

// GetCode -
func (sections Sections) GetCode() ([]byte, error) {
	return []byte(fmt.Sprintf(`[{
		"prim": "parameter",
		"args": [%s]
	},{
		"prim": "storage",
		"args": [%s]
	},{
		"prim": "code",
		"args": [%s]
	}]`, string(sections.Parameter), string(sections.ReturnType), string(sections.Code))), nil
}

// Execute -
func Execute(ctx context.Context, rpc noderpc.INode, event Event, args Args) ([]tokenbalance.TokenBalance, error) {
	parameter := event.Normalize(args.Parameters)
	if parameter == nil {
		logger.Warning().Msgf("%s event failed", args.Network)
		return nil, nil
	}
	storage := []byte(`[]`)
	code, err := event.GetCode()
	if err != nil {
		return nil, err
	}

	response, err := rpc.RunCode(ctx, code, storage, parameter, args.ChainID, args.Source, args.Initiator, args.Entrypoint, args.Protocol, args.Amount, args.HardGasLimitPerOperation)
	if err != nil {
		return nil, err
	}

	return event.Parse(response), nil
}

// NormalizeName -
func NormalizeName(name string) string {
	name = strings.ToLower(name)
	name = strings.ReplaceAll(name, "-", "")
	return strings.ReplaceAll(name, "_", "")
}
