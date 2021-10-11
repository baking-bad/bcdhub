package tokenbalance

import (
	"testing"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
)

func Test_singleAssetBalanceParser_Parse(t *testing.T) {
	tests := []struct {
		name string
		args string
		want []TokenBalance
	}{
		{
			name: "test 1",
			args: `[{"prim":"Elt","args": [{"string": "test"}, {"int": "100000000000000"}]}]`,
			want: []TokenBalance{
				{
					Address: "test",
					Value:   decimal.RequireFromString("100000000000000"),
				},
			},
		}, {
			name: "test 2",
			args: `[{"prim":"Elt","args": [{"bytes": "0000c67788ea8ada32b2426e1b02b9ebebdc2dc51007"}, {"int": "1000000000000000"}]}]`,
			want: []TokenBalance{
				{
					Address: "tz1djRgXXWWJiY1rpMECCxr5d9ZBqWewuiU1",
					Value:   decimal.RequireFromString("1000000000000000"),
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewSingleAssetBalance().Parse([]byte(tt.args))
			if err != nil {
				t.Errorf("Parse error=%s", err)
				return
			}
			assert.Equal(t, got, tt.want)
		})
	}
}

func Test_multiAssetBalanceParser_Parse(t *testing.T) {
	tests := []struct {
		name string
		args string
		want []TokenBalance
	}{
		{
			name: "test 1",
			args: `[{"prim":"Elt","args": [{"args": [{"string": "test"}, {"int": "1"}]}, {"int": "1000000000000000"}]}]`,
			want: []TokenBalance{
				{
					Address: "test",
					TokenID: 1,
					Value:   decimal.RequireFromString("1000000000000000"),
				},
			},
		}, {
			name: "test 2",
			args: `[{"prim":"Elt","args": [{"args": [{"bytes": "0000c67788ea8ada32b2426e1b02b9ebebdc2dc51007"}, {"int": "1"}]}, {"int": "1000000000000000"}]}]`,
			want: []TokenBalance{
				{
					Address: "tz1djRgXXWWJiY1rpMECCxr5d9ZBqWewuiU1",
					TokenID: 1,
					Value:   decimal.RequireFromString("1000000000000000"),
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewMultiAssetUpdate().Parse([]byte(tt.args))
			if err != nil {
				t.Errorf("Parse error=%v", err)
				return
			}
			assert.Equal(t, got, tt.want)
		})
	}
}

func Test_nftParser_Parse(t *testing.T) {
	tests := []struct {
		name string
		args string
		want []TokenBalance
	}{
		{
			name: "test 1",
			args: `[{"prim":"Elt","args": [{"int": "1"},{"string": "KT1BYYLfMjufYwqFtTSYJND7bzKNyK7mjrjM"}]}]`,
			want: []TokenBalance{
				{
					Address:        "KT1BYYLfMjufYwqFtTSYJND7bzKNyK7mjrjM",
					TokenID:        1,
					Value:          decimal.RequireFromString("1"),
					IsExclusiveNFT: true,
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewNftAsset().Parse([]byte(tt.args))
			if err != nil {
				t.Errorf("Parse error=%v", err)
				return
			}
			assert.Equal(t, got, tt.want)
		})
	}
}

func Test_nftOptionParser_Parse(t *testing.T) {
	tests := []struct {
		name string
		args string
		want []TokenBalance
	}{
		{
			name: "test 1",
			args: `[{"prim":"Elt","args": [{"int": "1"},{"prim": "Some", "args":[{"string": "KT1BYYLfMjufYwqFtTSYJND7bzKNyK7mjrjM"}]}]}]`,
			want: []TokenBalance{
				{
					Address:        "KT1BYYLfMjufYwqFtTSYJND7bzKNyK7mjrjM",
					TokenID:        1,
					Value:          decimal.RequireFromString("1"),
					IsExclusiveNFT: true,
				},
			},
		}, {
			name: "test 2",
			args: `[{"prim":"Elt","args": [{"int": "1"},{"prim": "None"}]}]`,
			want: []TokenBalance{
				{
					Address:        "",
					TokenID:        1,
					Value:          decimal.RequireFromString("0"),
					IsExclusiveNFT: true,
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewNftAssetOption().Parse([]byte(tt.args))
			if err != nil {
				t.Errorf("Parse error=%v", err)
				return
			}
			assert.Equal(t, got, tt.want)
		})
	}
}
