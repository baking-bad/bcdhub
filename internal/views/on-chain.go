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

// Return -
func (onChain *OnChain) Return() []byte {
	return onChain.ReturnType
}

// Execute -
func (onChain *OnChain) Execute(ctx context.Context, rpc noderpc.INode, args Args) ([]byte, error) {
	response, err := rpc.RunScriptView(ctx, noderpc.RunScriptViewRequest{
		ChainID:       args.ChainID,
		Contract:      args.Contract,
		View:          onChain.ViewName(),
		Input:         stdJSON.RawMessage(args.Parameters),
		Source:        args.Source,
		Payer:         args.Initiator,
		Gas:           args.HardGasLimitPerOperation,
		UnparsingMode: noderpc.UnparsingModeReadable,
	})
	if err != nil {
		return nil, err
	}

	var buf bytes.Buffer
	buf.WriteString(`{"args":[`)
	buf.Write(response)
	buf.WriteByte(']')
	buf.WriteByte('}')

	return buf.Bytes(), nil
}
