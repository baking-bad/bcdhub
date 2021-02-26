package views

import (
	"bytes"

	"github.com/baking-bad/bcdhub/internal/models/tzip"
)

// MichelsonStorageView -
type MichelsonStorageView struct {
	Parameter  []byte
	Code       []byte
	ReturnType []byte
	Name       string
}

// NewMichelsonStorageView -
func NewMichelsonStorageView(impl tzip.ViewImplementation, name string) *MichelsonStorageView {
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

// GetCode -
func (msv *MichelsonStorageView) GetCode(storageType []byte) ([]byte, error) {
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

// Parse -
func (msv *MichelsonStorageView) Parse(response []byte, output interface{}) error {
	return nil
}

// GetParameter -
func (msv *MichelsonStorageView) GetParameter(parameter string, storageValue []byte) ([]byte, error) {
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
