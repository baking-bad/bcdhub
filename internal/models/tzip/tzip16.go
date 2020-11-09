package tzip

import (
	"encoding/json"
	"fmt"

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

// License -
type License struct {
	Name    string `json:"name"`
	Details string `json:"details,omitempty"`
}

// UnmarshalJSON -
func (license *License) UnmarshalJSON(data []byte) error {
	switch data[0] {
	case '"':
		if err := json.Unmarshal(data, &license.Name); err != nil {
			return err
		}
	case '{':
		var buf struct {
			Name    string `json:"name"`
			Details string `json:"details,omitempty"`
		}
		if err := json.Unmarshal(data, &buf); err != nil {
			return err
		}
		license.Name = buf.Name
		license.Details = buf.Details
	}
	return nil
}

// Event -
type Event struct {
	Name            string                `json:"name"`
	Description     string                `json:"description"`
	Pure            string                `json:"pure"`
	Implementations []EventImplementation `json:"implementations"`
}

// EventImplementation -
type EventImplementation struct {
	MichelsonParameterView MichelsonParameterView `json:"michelson-parameter-view"`
}

// MichelsonParameterView -
type MichelsonParameterView struct {
	Parameter   interface{} `json:"parameter"`
	ReturnType  interface{} `json:"return-type"`
	Code        interface{} `json:"code"`
	Entrypoints []string    `json:"entrypoints"`
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
