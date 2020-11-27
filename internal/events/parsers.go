package events

import (
	"github.com/pkg/errors"
	"github.com/tidwall/gjson"
)

var parsers = map[string]Parser{
	SingleAssetBalanceUpdates: newSingleAssetBalanceParser(),
	MultiAssetBalanceUpdates:  newMultiAssetBalanceParser(),
}

// Parser -
type Parser interface {
	GetReturnType() gjson.Result
	Parse(gjson.Result) []TokenBalance
}

// GetParser -
func GetParser(name string, returnType []byte) (Parser, error) {
	p, ok := parsers[name]
	if !ok {
		return nil, errors.Errorf("Unknown event: %s", name)
	}
	typ := gjson.ParseBytes(returnType)
	if !isType(typ, p.GetReturnType()) {
		return nil, errors.Errorf("Invalid parser`s return type: %s", name)
	}
	return p, nil
}

func isType(a, b gjson.Result) bool {
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
			if !isType(arrA[i], arrB[i]) {
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

type singleAssetBalanceParser struct {
	ReturnType gjson.Result
}

func newSingleAssetBalanceParser() singleAssetBalanceParser {
	return singleAssetBalanceParser{
		ReturnType: gjson.Parse(`{ "prim": "map", "args": [ { "prim": "address"}, {"prim": "int"} ] }`),
	}
}

func (p singleAssetBalanceParser) GetReturnType() gjson.Result {
	return p.ReturnType
}

func (p singleAssetBalanceParser) Parse(response gjson.Result) []TokenBalance {
	balances := make([]TokenBalance, 0)
	for _, item := range response.Get("storage").Array() {
		balances = append(balances, TokenBalance{
			Address: item.Get("args.0.string").String(),
			Value:   item.Get("args.1.int").Int(),
		})
	}
	return balances
}

type multiAssetBalanceParser struct {
	ReturnType gjson.Result
}

func newMultiAssetBalanceParser() multiAssetBalanceParser {
	return multiAssetBalanceParser{
		ReturnType: gjson.Parse(`{ "prim": "map", "args": [ { "prim": "pair", "args": [{ "prim": "address"}, {"prim": "nat"}] }, { "prim" : "int" } ] }`),
	}
}

func (p multiAssetBalanceParser) GetReturnType() gjson.Result {
	return p.ReturnType
}

func (p multiAssetBalanceParser) Parse(response gjson.Result) []TokenBalance {
	balances := make([]TokenBalance, 0)
	for _, item := range response.Get("storage").Array() {
		balances = append(balances, TokenBalance{
			Value:   item.Get("args.1.int").Int(),
			Address: item.Get("args.0.args.0.string").String(),
			TokenID: item.Get("args.0.args.1.int").Int(),
		})
	}
	return balances
}
