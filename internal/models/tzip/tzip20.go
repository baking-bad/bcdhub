package tzip

import (
	"encoding/json"
	"fmt"

	"github.com/pkg/errors"
	"github.com/tidwall/gjson"
)

// TZIP20 -
type TZIP20 struct {
	Events []Event `json:"events,omitempty"`
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
	MichelsonParameterEvent       MichelsonParameterEvent `json:"michelson-parameter-view"`
	MichelsonInitialStorageEvent  Sections                `json:"michelson-initial-storage-event"`
	MichelsonExtendedStorageEvent Sections                `json:"michelson-extended-storage-event"`
}

// MichelsonParameterEvent -
type MichelsonParameterEvent struct {
	Sections
	Entrypoints []string `json:"entrypoints"`
}

// GetParser -
func (view MichelsonParameterEvent) GetParser() (BalanceViewParser, error) {
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

// Sections -
type Sections struct {
	Parameter  interface{} `json:"parameter"`
	ReturnType interface{} `json:"return-type"`
	Code       interface{} `json:"code"`
}

// CodeJSON -
func (sections Sections) CodeJSON() (gjson.Result, error) {
	parameter, err := json.Marshal(sections.Parameter)
	if err != nil {
		return gjson.Result{}, err
	}
	storage, err := json.Marshal(sections.ReturnType)
	if err != nil {
		return gjson.Result{}, err
	}
	code, err := json.Marshal(sections.Code)
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
