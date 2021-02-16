package views

import (
	"strings"

	"github.com/baking-bad/bcdhub/internal/models/tzip"
	"github.com/tidwall/gjson"
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
func (msv *MichelsonStorageView) GetCode(storageType gjson.Result) (gjson.Result, error) {
	var script strings.Builder
	script.WriteString(`[{"prim":"parameter","args":[`)
	if msv.Parameter != nil {
		script.WriteString(`{"prim":"pair","args":[`)
		script.Write(msv.Parameter)
		script.WriteString(",")
		script.WriteString(storageType.String())
		script.WriteString("]}")
	} else {
		script.WriteString(storageType.String())
	}
	script.WriteString(`]},{"prim":"storage","args":[{"prim":"option","args":[`)
	script.Write(msv.ReturnType)
	script.WriteString(`]}]},{"prim":"code","args":[[{"prim":"CAR"},`)
	script.Write(msv.Code)
	script.WriteString(`,{"prim":"SOME"},{"prim":"NIL","args":[{"prim":"operation"}]},{"prim":"PAIR"}]]}]`)
	return gjson.Parse(script.String()), nil
}

// Parse -
func (msv *MichelsonStorageView) Parse(response gjson.Result, output interface{}) error {
	return nil
}

// GetParameter -
func (msv *MichelsonStorageView) GetParameter(parameter string, storageValue gjson.Result) (gjson.Result, error) {
	var script strings.Builder
	if msv.Parameter != nil {
		script.WriteString(`{"prim":"Pair","args":[`)
		script.WriteString(parameter)
		script.WriteString(",")
		script.WriteString(storageValue.String())
		script.WriteString(`]}`)
	} else {
		script.WriteString(storageValue.String())
	}
	return gjson.Parse(script.String()), nil
}
