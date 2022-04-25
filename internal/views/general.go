package views

import (
	"context"
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

// Args -
type Args struct {
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
func Execute(ctx context.Context, rpc noderpc.INode, view View, args Args, output interface{}) error {
	response, err := ExecuteWithoutParsing(ctx, rpc, view, args)
	if err != nil {
		return err
	}

	return view.Parse(response, output)
}

// ExecuteWithoutParsing -
func ExecuteWithoutParsing(ctx context.Context, rpc noderpc.INode, view View, args Args) ([]byte, error) {
	script, err := rpc.GetScriptJSON(ctx, args.Contract, 0)
	if err != nil {
		return nil, err
	}

	parameter, err := view.GetParameter(args.Parameters, script.Storage)
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

	response, err := rpc.RunCode(ctx, code, storage, parameter, args.ChainID, args.Source, args.Initiator, args.Entrypoint, args.Protocol, args.Amount, args.HardGasLimitPerOperation)
	if err != nil {
		return nil, err
	}

	return response.Storage, nil
}
