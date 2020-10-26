package tzip

import (
	"encoding/json"
	"fmt"

	"github.com/baking-bad/bcdhub/internal/models/utils"
	"github.com/pkg/errors"
	"github.com/tidwall/gjson"
)

// TZIP16 -
type TZIP16 struct {
	Name        string   `json:"name,omitempty"`
	Description string   `json:"description,omitempty"`
	Version     string   `json:"version,omitempty"`
	License     *License `json:"license,omitempty"`
	Homepage    string   `json:"homepage,omitempty"`
	Authors     []string `json:"authors,omitempty"`
	Interfaces  []string `json:"interfaces,omitempty"`
	Events      []Event  `json:"events,omitempty"`
}

// ParseElasticJSON -
func (t *TZIP16) ParseElasticJSON(resp gjson.Result) {
	t.Name = resp.Get("_source.name").String()
	t.Description = resp.Get("_source.description").String()
	t.Version = resp.Get("_source.version").String()
	t.Homepage = resp.Get("_source.homepage").String()

	t.Authors = utils.StringArray(resp, "_source.authors")
	t.Interfaces = utils.StringArray(resp, "_source.interfaces")
	t.Events = make([]Event, 0)
	for _, hit := range resp.Get("_source.events").Array() {
		var event Event
		event.ParseElasticJSON(hit)
	}

	license := resp.Get("_source.license")
	if license.Exists() {
		t.License = &License{}
		t.License.ParseElasticJSON(license)
	}
}

// License -
type License struct {
	Name    string `json:"name"`
	Details string `json:"details,omitempty"`
}

// ParseElasticJSON -
func (license *License) ParseElasticJSON(resp gjson.Result) {
	license.Name = resp.Get("name").String()
	license.Details = resp.Get("details").String()
}

// Event -
type Event struct {
	Name            string                `json:"name"`
	Description     string                `json:"description"`
	Pure            string                `json:"pure"`
	Implementations []EventImplementation `json:"implementations"`
}

// ParseElasticJSON -
func (t *Event) ParseElasticJSON(resp gjson.Result) {
	t.Name = resp.Get("name").String()
	t.Description = resp.Get("description").String()
	t.Pure = resp.Get("pure").String()
	t.Implementations = make([]EventImplementation, 0)
	for _, hit := range resp.Get("implementations").Array() {
		var impl EventImplementation
		impl.ParseElasticJSON(hit)
		t.Implementations = append(t.Implementations, impl)
	}
}

// EventImplementation -
type EventImplementation struct {
	MichelsonParameterView MichelsonParameterView `json:"michelson-parameter-view"`
}

// ParseElasticJSON -
func (t *EventImplementation) ParseElasticJSON(resp gjson.Result) {
	t.MichelsonParameterView.ParseElasticJSON(resp.Get("michelson-parameter-view"))
}

// MichelsonParameterView -
type MichelsonParameterView struct {
	Parameter   interface{} `json:"parameter"`
	ReturnType  interface{} `json:"return-type"`
	Code        interface{} `json:"code"`
	Entrypoints []string    `json:"entrypoints"`
}

// ParseElasticJSON -
func (view *MichelsonParameterView) ParseElasticJSON(resp gjson.Result) {
	view.Parameter = resp.Get("parameter").Value()
	view.ReturnType = resp.Get("return-type").Value()
	view.Code = resp.Get("code").Value()
	view.Entrypoints = utils.StringArray(resp, "entrypoints")
}

// CodeJSON -
func (view MichelsonParameterView) CodeJSON() (gjson.Result, error) {
	parameter, err := json.Marshal(view.Parameter)
	if err != nil {
		return gjson.Result{}, err
	}
	storage, err := json.Marshal(view.ReturnType)
	if err != nil {
		return gjson.Result{}, err
	}
	code, err := json.Marshal(view.Code)
	if err != nil {
		return gjson.Result{}, err
	}

	return gjson.Parse(fmt.Sprintf(`[{
		"prim": "parameter",
		"args": [%s]
	},{
		"prim": "storage",
		"args": [%s]
	},{
		"prim": "code",
		"args": [%s]
	}]`, string(parameter), string(storage), string(code))), nil
}

// GetParser -
func (view MichelsonParameterView) GetParser() (BalanceViewParser, error) {
	b, err := json.Marshal(view.ReturnType)
	if err != nil {
		return nil, err
	}
	typ := gjson.ParseBytes(b)
	for _, p := range []BalanceViewParser{
		onlyBalanceParser{},
		withTokenIDBalanceParser{},
	} {
		if compareTypes(typ, p.GetReturnType()) {
			return p, nil
		}
	}
	return nil, errors.Errorf("Unknown balance parser")
}

func compareTypes(a, b gjson.Result) bool {
	primA, primB, ok := getKey(a, b, "prim")
	if !ok || primA.String() != primB.String() {
		return false
	}

	if argsA, argsB, argOk := getKey(a, b, "args"); argOk {
		arrA := argsA.Array()
		arrB := argsB.Array()
		if len(arrA) != len(arrB) {
			return false
		}

		for i := range arrA {
			if !compareTypes(arrA[i], arrB[i]) {
				return false
			}
		}
	}

	return true
}

func getKey(a, b gjson.Result, key string) (gjson.Result, gjson.Result, bool) {
	keyA := a.Get(key)
	keyB := b.Get(key)

	if keyA.Exists() != keyB.Exists() {
		return a, b, false
	}
	return keyA, keyB, keyA.Exists()
}
