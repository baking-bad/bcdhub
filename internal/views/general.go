package views

import (
	"strings"

	"github.com/baking-bad/bcdhub/internal/noderpc"
	"github.com/pkg/errors"
	"github.com/tidwall/gjson"
)

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
	GetCode(storageType gjson.Result) (gjson.Result, error)
	GetParameter(parameter string, storageType gjson.Result) (gjson.Result, error)
	Parse(response gjson.Result, output interface{}) error
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
func ExecuteWithoutParsing(rpc noderpc.INode, view View, ctx Context) (gjson.Result, error) {
	script, err := rpc.GetScriptJSON(ctx.Contract, 0)
	if err != nil {
		return gjson.Result{}, err
	}

	storageValue := script.Get(`storage`)
	parameter, err := view.GetParameter(ctx.Parameters, storageValue)
	if err != nil {
		return gjson.Result{}, err
	}

	storageType := script.Get(`code.#(prim=="storage").args.0`)
	code, err := view.GetCode(storageType)
	if err != nil {
		return gjson.Result{}, err
	}

	storage := gjson.Parse(`{"prim": "None"}`)

	response, err := rpc.RunCode(code, storage, parameter, ctx.ChainID, ctx.Source, ctx.Initiator, ctx.Entrypoint, ctx.Protocol, ctx.Amount, ctx.HardGasLimitPerOperation)
	if err != nil {
		return gjson.Result{}, err
	}
	if err := checkResponseError(response); err != nil {
		return gjson.Result{}, err
	}

	return response.Get("storage"), nil
}

func checkResponseError(response gjson.Result) error {
	if !response.IsArray() {
		return nil
	}

	var builder strings.Builder
	for i, item := range response.Array() {
		if i > 0 {
			if err := builder.WriteByte('\n'); err != nil {
				return err
			}
		}
		if _, err := builder.WriteString(item.Get("id").String()); err != nil {
			return err
		}
	}
	return errors.Wrap(ErrNodeReturn, builder.String())
}
