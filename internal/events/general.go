package events

import (
	stdJSON "encoding/json"
	"fmt"
	"strings"

	"github.com/baking-bad/bcdhub/internal/noderpc"
	"github.com/baking-bad/bcdhub/internal/parsers/tokenbalance"
)

// Event -
type Event interface {
	GetCode() ([]byte, error)
	Parse(response noderpc.RunCodeResponse) []tokenbalance.TokenBalance
	Normalize(parameter string) []byte
}

// Context -
type Context struct {
	Network                  string
	Protocol                 string
	Parameters               string
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
func Execute(rpc noderpc.INode, event Event, ctx Context) ([]tokenbalance.TokenBalance, error) {
	parameter := event.Normalize(ctx.Parameters)
	storage := []byte(`[]`)
	code, err := event.GetCode()
	if err != nil {
		return nil, err
	}

	response, err := rpc.RunCode(code, storage, parameter, ctx.ChainID, ctx.Source, ctx.Initiator, ctx.Entrypoint, ctx.Protocol, ctx.Amount, ctx.HardGasLimitPerOperation)
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
