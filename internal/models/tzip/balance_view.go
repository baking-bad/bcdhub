package tzip

import "github.com/tidwall/gjson"

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
