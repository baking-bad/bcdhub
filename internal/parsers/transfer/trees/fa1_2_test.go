package trees

import (
	"math/big"
	"testing"

	"github.com/baking-bad/bcdhub/internal/bcd/ast"
	"github.com/baking-bad/bcdhub/internal/bcd/types"
	"github.com/baking-bad/bcdhub/internal/models/operation"
	"github.com/baking-bad/bcdhub/internal/models/transfer"
	"github.com/stretchr/testify/assert"
)

func TestMakeFa1_2Transfers(t *testing.T) {
	tests := []struct {
		name      string
		tree      *ast.TypedAst
		operation operation.Operation
		want      []*transfer.Transfer
		wantErr   bool
	}{
		{
			name: "FA 1.2",
			tree: GetFA1_2Transfer(),
			operation: operation.Operation{
				Network:    "edo2net",
				Parameters: `{"entrypoint":"transfer","value":{"prim":"Pair","args":[{"string":"tz1grSQDByRpnVs7sPtaprNZRp531ZKz6Jmm"},{"string":"tz1TGu6TN5GSez2ndXXeDX6LgUDvLzPLqgYV"},{"int":"100"}]}}`,
			},
			want: []*transfer.Transfer{
				{
					Network:      "edo2net",
					From:         "tz1grSQDByRpnVs7sPtaprNZRp531ZKz6Jmm",
					To:           "tz1TGu6TN5GSez2ndXXeDX6LgUDvLzPLqgYV",
					AmountBigInt: big.NewInt(100),
				},
			},
		}, {
			name: "test 2",
			tree: GetFA1_2Transfer(),
			operation: operation.Operation{
				Network:    "mainnet",
				Parameters: "{\"entrypoint\":\"transfer\",\"value\":{\"prim\":\"Pair\",\"args\":[{\"bytes\":\"011871cfab6dafee00330602b4342b6500c874c93b00\"},{\"prim\":\"Pair\",\"args\":[{\"bytes\":\"0000c2473c617946ce7b9f6843f193401203851cb2ec\"},{\"int\":\"7874880\"}]}]}}",
			},
			want: []*transfer.Transfer{
				{
					Network:      "mainnet",
					From:         "KT1Ap287P1NzsnToSJdA4aqSNjPomRaHBZSr",
					To:           "tz1dMH7tW7RhdvVMR4wKVFF1Ke8m8ZDvrTTE",
					AmountBigInt: big.NewInt(7.87488e+06),
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			params := types.NewParameters([]byte(tt.operation.Parameters))
			if _, err := tt.tree.FromParameters(params); err != nil {
				t.Errorf("FromParameters() error = %v", err)
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
