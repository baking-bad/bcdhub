package tokenbalance

import (
	"fmt"
	"math/big"

	"github.com/baking-bad/bcdhub/internal/bcd/forge"
	"github.com/tidwall/gjson"
)

// MultiAsset -
type MultiAsset struct {
	ReturnType gjson.Result
}

// NewMultiAssetBalance -
func NewMultiAssetBalance() MultiAsset {
	return MultiAsset{
		ReturnType: gjson.Parse(`{ "prim": "map", "args": [ { "prim": "pair", "args": [{ "prim": "address"}, {"prim": "nat"}] }, { "prim" : "nat" } ] }`),
	}
}

// NewMultiAssetUpdate -
func NewMultiAssetUpdate() MultiAsset {
	return MultiAsset{
		ReturnType: gjson.Parse(`{ "prim": "map", "args": [ { "prim": "pair", "args": [{ "prim": "address"}, {"prim": "nat"}] }, { "prim" : "int" } ] }`),
	}
}

// GetReturnType -
func (p MultiAsset) GetReturnType() gjson.Result {
	return p.ReturnType
}

// Parse -
func (p MultiAsset) Parse(item gjson.Result) (TokenBalance, error) {
	balance := big.NewInt(0)
	if _, ok := balance.SetString(item.Get("args.1.int").String(), 10); !ok {
		return TokenBalance{}, fmt.Errorf("Invalid int in parsing multi-asset balance: %s", item.Raw)
	}
	var address string
	switch {
	case item.Get("args.0.args.0.string").Exists():
		address = item.Get("args.0.args.0.string").String()
	case item.Get("args.0.args.0.bytes").Exists():
		val, err := forge.UnforgeAddress(item.Get("args.0.args.0.bytes").String())
		if err != nil {
			return TokenBalance{}, err
		}
		address = val
	}
	return TokenBalance{
		Value:   balance,
		Address: address,
		TokenID: item.Get("args.0.args.1.int").Int(),
	}, nil
}
