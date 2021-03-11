package views

import (
	"strings"

	jsoniter "github.com/json-iterator/go"

	"github.com/baking-bad/bcdhub/internal/noderpc"
	"github.com/pkg/errors"
)

var json = jsoniter.ConfigCompatibleWithStandardLibrary

// errors
var (
	ErrNodeReturn = errors.New(`Node return error`)
)

// Context -
type Context struct {
	Network                  string
	Protocol                 string
	Contract                 string
	Parameters               string
	Source                   string
	Initiator                string
	Entrypoint               string
	ChainID                  string
	HardGasLimitPerOperation int64
	Amount                   int64
}

// View -
type View interface {
	GetCode(storageType []byte) ([]byte, error)
	GetParameter(parameter string, storageType []byte) ([]byte, error)
	Parse(response []byte, output interface{}) error
}

// NormalizeName -
func NormalizeName(name string) string {
	name = strings.ToLower(name)
	name = strings.ReplaceAll(name, "-", "")
	return strings.ReplaceAll(name, "_", "")
}

// Execute -
func Execute(rpc noderpc.INode, view View, ctx Context, output interface{}) error {
	response, err := ExecuteWithoutParsing(rpc, view, ctx)
	if err != nil {
		return err
	}

	return view.Parse(response, output)
}

// ExecuteWithoutParsing -
func ExecuteWithoutParsing(rpc noderpc.INode, view View, ctx Context) ([]byte, error) {
	script, err := rpc.GetScriptJSON(ctx.Contract, 0)
	if err != nil {
		return nil, err
	}

	parameter, err := view.GetParameter(ctx.Parameters, script.Storage)
	if err != nil {
		return nil, err
	}

	storageType, err := json.Marshal(script.Code.Storage[0])
	if err != nil {
		return nil, err
	}
	code, err := view.GetCode(storageType)
	if err != nil {
		return nil, err
	}

	storage := []byte(`{"prim": "None"}`)

	response, err := rpc.RunCode(code, storage, parameter, ctx.ChainID, ctx.Source, ctx.Initiator, ctx.Entrypoint, ctx.Protocol, ctx.Amount, ctx.HardGasLimitPerOperation)
	if err != nil {
		return nil, err
	}

	return response.Storage, nil
}
