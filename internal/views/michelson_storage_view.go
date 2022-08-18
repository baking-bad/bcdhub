package views

import (
	"bytes"
	"context"

	"github.com/baking-bad/bcdhub/internal/models/contract"
	"github.com/baking-bad/bcdhub/internal/noderpc"
)

// MichelsonStorageView -
type MichelsonStorageView struct {
	Parameter  []byte
	Code       []byte
	ReturnType []byte
	Name       string
}

// NewMichelsonStorageView -
func NewMichelsonStorageView(impl contract.ViewImplementation, name string) *MichelsonStorageView {
	var parameter []byte
	if !impl.MichelsonStorageView.IsParameterEmpty() {
		parameter = impl.MichelsonStorageView.Parameter
	}
	return &MichelsonStorageView{
		Parameter:  parameter,
		ReturnType: impl.MichelsonStorageView.ReturnType,
		Code:       impl.MichelsonStorageView.Code,
		Name:       name,
	}
}

func (msv *MichelsonStorageView) buildCode(storageType []byte) ([]byte, error) {
	var script bytes.Buffer
	script.WriteString(`[{"prim":"parameter","args":[`)
	if msv.Parameter != nil {
		script.WriteString(`{"prim":"pair","args":[`)
		script.Write(msv.Parameter)
		script.WriteString(",")
		if _, err := script.Write(storageType); err != nil {
			return nil, err
		}
		script.WriteString("]}")
	} else if _, err := script.Write(storageType); err != nil {
		return nil, err
	}
	script.WriteString(`]},{"prim":"storage","args":[{"prim":"option","args":[`)
	script.Write(msv.ReturnType)
	script.WriteString(`]}]},{"prim":"code","args":[[{"prim":"CAR"},`)
	script.Write(msv.Code)
	script.WriteString(`,{"prim":"SOME"},{"prim":"NIL","args":[{"prim":"operation"}]},{"prim":"PAIR"}]]}]`)
	return script.Bytes(), nil
}

func (msv *MichelsonStorageView) buildParameter(_, parameter string, storageValue []byte) ([]byte, error) {
	var script bytes.Buffer
	if msv.Parameter != nil {
		script.WriteString(`{"prim":"Pair","args":[`)
		script.WriteString(parameter)
		script.WriteString(",")
		if _, err := script.Write(storageValue); err != nil {
			return nil, err
		}
		script.WriteString(`]}`)
	} else if _, err := script.Write(storageValue); err != nil {
		return nil, err
	}
	return script.Bytes(), nil
}

// Return -
func (msv *MichelsonStorageView) Return() []byte {
	return msv.ReturnType
}

// Execute -
func (msv *MichelsonStorageView) Execute(ctx context.Context, rpc noderpc.INode, args Args) ([]byte, error) {
	script, err := rpc.GetScriptJSON(ctx, args.Contract, 0)
	if err != nil {
		return nil, err
	}

	parameter, err := msv.buildParameter(args.Contract, args.Parameters, script.Storage)
	if err != nil {
		return nil, err
	}

	storageType, err := json.Marshal(script.Code.Storage[0])
	if err != nil {
		return nil, err
	}
	code, err := msv.buildCode(storageType)
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
