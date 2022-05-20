package views

import (
	"bytes"
	"context"
	stdJSON "encoding/json"

	"github.com/baking-bad/bcdhub/internal/bcd/consts"
	"github.com/baking-bad/bcdhub/internal/noderpc"
	"github.com/pkg/errors"
)

// OnChain -
type OnChain struct {
	Name       Name
	Parameter  stdJSON.RawMessage
	ReturnType stdJSON.RawMessage
	Code       stdJSON.RawMessage
}

// ViewName -
func (onChain OnChain) ViewName() string {
	return onChain.Name.Value
}

// Name -
type Name struct {
	Value string `json:"string"`
}

// UnmarshalJSON -
func (onChain *OnChain) UnmarshalJSON(data []byte) error {
	buf := struct {
		Args []any  `json:"args"`
		Prim string `json:"prim"`
	}{
		Args: []any{&onChain.Name, &onChain.Parameter, &onChain.ReturnType, &onChain.Code},
	}
	if err := json.Unmarshal(data, &buf); err != nil {
		return err
	}
	if buf.Prim != consts.View {
		return errors.Errorf("invalid primitive '%s', 'view' is expected", buf.Prim)
	}
	return nil
}

func (onChain *OnChain) buildCode(storageType []byte) ([]byte, error) {
	var script bytes.Buffer
	script.WriteString(`[{"prim":"parameter","args":[{"prim":"pair","args":[`)
	script.Write(onChain.Parameter)
	script.WriteString(`,{"prim":"address"}]}]},{"prim":"storage","args":[{"prim":"option","args":[`)
	script.Write(onChain.ReturnType)
	script.WriteString(`]}]},{"prim":"code","args":[[{"prim":"CAR"},{"prim":"UNPAIR"},{"prim":"VIEW","args":[{"string":"`)
	script.WriteString(onChain.ViewName())
	script.WriteString(`"},`)
	script.Write(onChain.ReturnType)
	script.WriteString(`]},{"prim":"NIL","args":[{"prim":"operation"}]},{"prim":"PAIR"}]]}]`)
	return script.Bytes(), nil
}

func (onChain *OnChain) buildParameter(address, parameter string, storageValue []byte) ([]byte, error) {
	var script bytes.Buffer
	script.WriteString(`{"prim":"Pair","args":[`)
	script.WriteString(parameter)
	script.WriteString(`,{"string":"`)
	script.WriteString(address)
	script.WriteString(`"}]}`)
	return script.Bytes(), nil
}

// Return -
func (onChain *OnChain) Return() []byte {
	return onChain.ReturnType
}

// Execute -
func (onChain *OnChain) Execute(ctx context.Context, rpc noderpc.INode, args Args) ([]byte, error) {
	parameter, err := onChain.buildParameter(args.Contract, args.Parameters, nil)
	if err != nil {
		return nil, err
	}

	code, err := onChain.buildCode(nil)
	if err != nil {
		return nil, err
	}

	storage := []byte(`{"prim": "None"}`)

	response, err := rpc.RunCode(ctx, code, storage, parameter, args.ChainID, args.Source, args.Initiator, "", args.Protocol, args.Amount, args.HardGasLimitPerOperation)
	if err != nil {
		return nil, err
	}

	return response.Storage, nil
}
