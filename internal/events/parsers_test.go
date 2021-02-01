package events

import (
	"math/big"
	"testing"

	"github.com/tidwall/gjson"
	"gopkg.in/go-playground/assert.v1"
)

func Test_singleAssetBalanceParser_Parse(t *testing.T) {
	tests := []struct {
		name string
		args string
		want []TokenBalance
	}{
		{
			name: "test 1",
			args: `{"storage": [{"args": [{"string": "test"}, {"int": "1000000000000000"}]}]}`,
			want: []TokenBalance{
				{
					Address: "test",
					Value:   newBigIntFromString("1000000000000000"),
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := singleAssetBalanceParser{}

			data := gjson.Parse(tt.args)
			got := p.Parse(data)
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
		want []TokenBalance
	}{
		{
			name: "test 1",
			args: `{"storage": [{"args": [{"args": [{"string": "test"}, {"int": "1"}]}, {"int": "1000000000000000"}]}]}`,
			want: []TokenBalance{
				{
					Address: "test",
					TokenID: 1,
					Value:   newBigIntFromString("1000000000000000"),
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := multiAssetBalanceParser{}
			data := gjson.Parse(tt.args)
			got := p.Parse(data)
			assert.Equal(t, got, tt.want)
		})
	}
}
