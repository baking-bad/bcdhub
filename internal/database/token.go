package database

import (
	"encoding/json"
	"fmt"

	"github.com/jinzhu/gorm"
	"github.com/jinzhu/gorm/dialects/postgres"
	"github.com/pkg/errors"
	"github.com/tidwall/gjson"
)

// Token -
type Token struct {
	ID       uint   `gorm:"primary_key" json:"-"`
	Name     string `json:"name"`
	Symbol   string `json:"symbol"`
	Decimals uint   `json:"decimals"`
	Contract string `json:"contract"`
	Network  string `json:"network"`
	TokenID  uint   `json:"token_id"`
	DAppID   uint   `json:"-"`

	MetadataJSON postgres.Jsonb `json:"-"`
	Metadata     TokenMetadata  `gorm:"-" json:"-"`
}

// BeforeSave -
func (token *Token) BeforeSave(tx *gorm.DB) error {
	return token.MetadataJSON.Scan(token.Metadata)
}

// AfterFind -
func (token *Token) AfterFind(tx *gorm.DB) error {
	b, err := token.MetadataJSON.MarshalJSON()
	if err != nil {
		return err
	}
	if string(b) == "null" {
		return nil
	}
	return json.Unmarshal(b, &token.Metadata)
}

// TokenMetadata -
type TokenMetadata struct {
	Version    string      `json:"version"`
	License    string      `json:"license"`
	Authors    []string    `json:"authors"`
	Interfaces []string    `json:"interfaces"`
	Views      []TokenView `json:"views"`
}

// TokenView -
type TokenView struct {
	Name            string                    `json:"name"`
	Description     string                    `json:"description"`
	Pure            string                    `json:"pure"`
	Implementations []TokenViewImplementation `json:"implementations"`
}

//TokenViewImplementation -
type TokenViewImplementation struct {
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

// GetTokens -
func (d *db) GetTokens() ([]Token, error) {
	var tokens []Token

	if err := d.Find(&tokens).Error; err != nil {
		return nil, err
	}

	return tokens, nil
}

// BalanceViewParser -
type BalanceViewParser interface {
	GetReturnType() gjson.Result
	Parse(gjson.Result) []TokenBalance
}

// TokenBalance -
type TokenBalance struct {
	Address string
	TokenID int64
	Value   int64
}

type onlyBalanceParser struct{}

func (p onlyBalanceParser) GetReturnType() gjson.Result {
	return gjson.Parse(`{ "prim": "map", "args": [ { "prim": "address"}, {"prim": "int"} ] }`)
}

func (p onlyBalanceParser) Parse(response gjson.Result) []TokenBalance {
	balances := make([]TokenBalance, 0)
	for _, item := range response.Get("storage").Array() {
		balances = append(balances, TokenBalance{
			Address: item.Get("args.0.string").String(),
			Value:   item.Get("args.1.int").Int(),
		})
	}
	return balances
}

type withTokenIDBalanceParser struct{}

func (p withTokenIDBalanceParser) GetReturnType() gjson.Result {
	return gjson.Parse(`{ "prim": "map", "args": [ { "prim": "pair", "args": [{ "prim": "address"}, {"prim": "nat"}] }, { "prim" : "int" } ] }`)
}

func (p withTokenIDBalanceParser) Parse(response gjson.Result) []TokenBalance {
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
