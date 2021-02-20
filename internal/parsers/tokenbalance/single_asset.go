package tokenbalance

import (
	"fmt"
	"math/big"

	"github.com/baking-bad/bcdhub/internal/bcd/forge"
	"github.com/tidwall/gjson"
)

// SingleAsset -
type SingleAsset struct {
	ReturnType gjson.Result
}

// NewSingleAssetBalance -
func NewSingleAssetBalance() SingleAsset {
	return SingleAsset{
		ReturnType: gjson.Parse(`{ "prim": "map", "args": [ { "prim": "address"}, {"prim": "nat"} ] }`),
	}
}

// NewSingleAssetUpdate -
func NewSingleAssetUpdate() SingleAsset {
	return SingleAsset{
		ReturnType: gjson.Parse(`{ "prim": "map", "args": [ { "prim": "address"}, {"prim": "int"} ] }`),
	}
}

// GetReturnType -
func (p SingleAsset) GetReturnType() gjson.Result {
	return p.ReturnType
}

// Parse -
func (p SingleAsset) Parse(item gjson.Result) (TokenBalance, error) {
	balance := big.NewInt(0)
	if _, ok := balance.SetString(item.Get("args.1.int").String(), 10); !ok {
		return TokenBalance{}, fmt.Errorf("Invalid int in parsing single-asset balance: %s", item.Raw)
	}
	var address string
	switch {
	case item.Get("args.0.string").Exists():
		address = item.Get("args.0.string").String()
	case item.Get("args.0.bytes").Exists():
		val, err := forge.UnforgeAddress(item.Get("args.0.bytes").String())
		if err != nil {
			return TokenBalance{}, err
		}
		address = val
	}
	return TokenBalance{
		Address: address,
		Value:   balance,
	}, nil
}
