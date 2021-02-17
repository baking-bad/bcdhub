package tokenbalance

import (
	"math/big"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/tidwall/gjson"
)

func Test_singleAssetBalanceParser_Parse(t *testing.T) {
	tests := []struct {
		name string
		args string
		want TokenBalance
	}{
		{
			name: "test 1",
			args: `{"args": [{"string": "test"}, {"int": "100000000000000"}]}`,
			want: TokenBalance{
				Address: "test",
				Value:   newBigIntFromString("100000000000000"),
			},
		}, {
			name: "test 2",
			args: `{"args": [{"bytes": "0000c67788ea8ada32b2426e1b02b9ebebdc2dc51007"}, {"int": "1000000000000000"}]}`,
			want: TokenBalance{
				Address: "tz1djRgXXWWJiY1rpMECCxr5d9ZBqWewuiU1",
				Value:   newBigIntFromString("1000000000000000"),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := SingleAsset{}

			data := gjson.Parse(tt.args)
			got, err := p.Parse(data)
			if err != nil {
				t.Errorf("Parse error=%s", err)
				return
			}
			assert.Equal(t, got, tt.want)
		})
	}
}

func newBigIntFromString(val string) *big.Int {
	i := big.NewInt(0)
	i, _ = i.SetString(val, 10)
	return i
}

func Test_multiAssetBalanceParser_Parse(t *testing.T) {
	tests := []struct {
		name string
		args string
		want TokenBalance
	}{
		{
			name: "test 1",
			args: `{"args": [{"args": [{"string": "test"}, {"int": "1"}]}, {"int": "1000000000000000"}]}`,
			want: TokenBalance{
				Address: "test",
				TokenID: 1,
				Value:   newBigIntFromString("1000000000000000"),
			},
		}, {
			name: "test 2",
			args: `{"args": [{"args": [{"bytes": "0000c67788ea8ada32b2426e1b02b9ebebdc2dc51007"}, {"int": "1"}]}, {"int": "1000000000000000"}]}`,
			want: TokenBalance{
				Address: "tz1djRgXXWWJiY1rpMECCxr5d9ZBqWewuiU1",
				TokenID: 1,
				Value:   newBigIntFromString("1000000000000000"),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := MultiAsset{}
			data := gjson.Parse(tt.args)
			got, err := p.Parse(data)
			if err != nil {
				t.Errorf("Parse error=%w", err)
				return
			}
			assert.Equal(t, got, tt.want)
		})
	}
}
