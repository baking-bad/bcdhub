package trees

import (
	"encoding/json"
	"testing"

	"github.com/baking-bad/bcdhub/internal/bcd/ast"
	"github.com/baking-bad/bcdhub/internal/bcd/base"
	"github.com/baking-bad/bcdhub/internal/bcd/types"
	"github.com/baking-bad/bcdhub/internal/models/operation"
	"github.com/baking-bad/bcdhub/internal/models/transfer"
	modelTypes "github.com/baking-bad/bcdhub/internal/models/types"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
)

func newDecimal(val string) decimal.Decimal {
	i, _ := decimal.NewFromString(val)
	return i
}

func TestMakeFa1_2Transfers(t *testing.T) {
	tests := []struct {
		name      string
		tree      ast.Node
		operation operation.Operation
		want      []*transfer.Transfer
		wantErr   bool
	}{
		{
			name: "FA 1.2",
			tree: GetFA1_2Transfer(),
			operation: operation.Operation{
				Network:    modelTypes.Edo2net,
				Parameters: []byte(`{"entrypoint":"transfer","value":{"prim":"Pair","args":[{"string":"tz1grSQDByRpnVs7sPtaprNZRp531ZKz6Jmm"},{"string":"tz1TGu6TN5GSez2ndXXeDX6LgUDvLzPLqgYV"},{"int":"100"}]}}`),
			},
			want: []*transfer.Transfer{
				{
					Network: modelTypes.Edo2net,
					From:    "tz1grSQDByRpnVs7sPtaprNZRp531ZKz6Jmm",
					To:      "tz1TGu6TN5GSez2ndXXeDX6LgUDvLzPLqgYV",
					Amount:  newDecimal("100"),
				},
			},
		}, {
			name: "test 2",
			tree: GetFA1_2Transfer(),
			operation: operation.Operation{
				Network:    modelTypes.Mainnet,
				Parameters: []byte("{\"entrypoint\":\"transfer\",\"value\":{\"prim\":\"Pair\",\"args\":[{\"bytes\":\"011871cfab6dafee00330602b4342b6500c874c93b00\"},{\"prim\":\"Pair\",\"args\":[{\"bytes\":\"0000c2473c617946ce7b9f6843f193401203851cb2ec\"},{\"int\":\"7874880\"}]}]}}"),
			},
			want: []*transfer.Transfer{
				{
					Network: modelTypes.Mainnet,
					From:    "KT1Ap287P1NzsnToSJdA4aqSNjPomRaHBZSr",
					To:      "tz1dMH7tW7RhdvVMR4wKVFF1Ke8m8ZDvrTTE",
					Amount:  newDecimal("7874880"),
				},
			},
		}, {
			name: "test 3",
			tree: GetFA1_2Transfer(),
			operation: operation.Operation{
				Network: modelTypes.Mainnet,
				Parameters: []byte(`{
					"entrypoint": "transfer",
					"value": {
					"prim": "Pair",
					"args": [
						{
						"bytes": "012d1c7c9c5add2d5161f70c19caa6aacd23cd570000"
						},
						{
						"prim": "Pair",
						"args": [
							{
							"bytes": "000018299ff2a891bc1fbedc15c0750183df1d0b8cb2"
							},
							{
							"int": "15019000009999999295"
							}
						]
						}
					]
					}
				}`),
			},
			want: []*transfer.Transfer{
				{
					Network: modelTypes.Mainnet,
					From:    "KT1ChJ6h8Crjdfds99DLpE5USynQTmCJtB3T",
					To:      "tz1Mqnms73LqgBCYiM7e5k12VyWNQG8ytcGb",
					Amount:  newDecimal("15019000009999999295"),
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			params := types.NewParameters(tt.operation.Parameters)
			node := new(base.Node)
			if err := json.Unmarshal(params.Value, node); err != nil {
				t.Errorf("Unmarshal() error = %v", err)
				return
			}

			if err := tt.tree.ParseValue(node); err != nil {
				t.Errorf("ParseValue() error = %v", err)
				return
			}
			got, err := MakeFa1_2Transfers(tt.tree, tt.operation)
			if (err != nil) != tt.wantErr {
				t.Errorf("MakeFa1_2Transfers() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !assert.Len(t, got, len(tt.want)) {
				return
			}
			for i := range tt.want {
				tt.want[i].ID = got[i].ID
				if !assert.Equal(t, tt.want[i], got[i]) {
					return
				}
			}
		})
	}
}
