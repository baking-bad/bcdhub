package tokenbalance

import (
	"math/big"
	"strings"

	"github.com/pkg/errors"
	"github.com/tidwall/gjson"
)

var parsers = map[string][]Parser{
	SingleAssetBalanceUpdates: {
		NewSingleAssetBalance(),
		NewSingleAssetUpdate(),
	},
	MultiAssetBalanceUpdates: {
		NewMultiAssetBalance(),
		NewMultiAssetUpdate(),
	},
}

// Parser -
type Parser interface {
	GetReturnType() gjson.Result
	Parse(item gjson.Result) (TokenBalance, error)
}

// TokenBalance -
type TokenBalance struct {
	Address string
	TokenID int64
	Value   *big.Int
}

// GetParser -
func GetParser(name string, returnType []byte) (Parser, error) {
	p, ok := parsers[NormalizeName(name)]
	if !ok {
		for _, ps := range parsers {
			item, err := findParser(ps, returnType)
			if err == nil {
				return item, nil
			}
		}
		return nil, errors.Wrap(ErrUnknownParser, name)
	}

	return findParser(p, returnType)
}

// GetParserForBigMap -
func GetParserForBigMap(returnType []byte) (Parser, error) {
	var s strings.Builder
	s.WriteString(`{"prim":"map","args":`)
	typ := gjson.ParseBytes(returnType)
	s.WriteString(typ.Get("args").Raw)
	s.WriteByte('}')
	return GetParser("", []byte(s.String()))
}

func findParser(p []Parser, returnType []byte) (Parser, error) {
	typ := gjson.ParseBytes(returnType)
	for i := range p {
		if isType(typ, p[i].GetReturnType()) {
			return p[i], nil
		}
	}
	return nil, errors.Errorf("Invalid parser`s return type: %s", string(returnType))
}

func getKey(a, b gjson.Result, key string) (gjson.Result, gjson.Result, bool) {
	keyA := a.Get(key)
	keyB := b.Get(key)

	if keyA.Exists() != keyB.Exists() {
		return a, b, false
	}
	return keyA, keyB, keyA.Exists()
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

// NormalizeName -
func NormalizeName(name string) string {
	name = strings.ToLower(name)
	name = strings.ReplaceAll(name, "-", "")
	return strings.ReplaceAll(name, "_", "")
}
