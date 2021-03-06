package trees

import (
	"encoding/json"
	"math/big"
	"testing"

	"github.com/baking-bad/bcdhub/internal/bcd/ast"
	"github.com/baking-bad/bcdhub/internal/bcd/base"
	"github.com/baking-bad/bcdhub/internal/bcd/types"
	"github.com/baking-bad/bcdhub/internal/models/operation"
	"github.com/baking-bad/bcdhub/internal/models/transfer"
	"github.com/stretchr/testify/assert"
)

func TestMakeFa2Transfers(t *testing.T) {
	tests := []struct {
		name      string
		tree      ast.Node
		operation operation.Operation
		want      []*transfer.Transfer
		wantErr   bool
	}{
		{
			name: "FA 2",
			tree: GetFA2Transfer(),
			operation: operation.Operation{
				Network:    "mainnet",
				Parameters: `{"entrypoint":"transfer","value":[{"prim":"Pair","args":[{"string":"tz1gHJt7J1aEtW2wpCR5RJd3CpnbVxUTaEXS"},[{"prim":"Pair","args":[{"string":"tz1gsJENNUwg7fQiRwQi5zJYaj7YtwwsE3y2"},{"prim":"Pair","args":[{"int":"0"},{"int":"1000000000"}]}]}]]}]}`,
			},
			want: []*transfer.Transfer{
				{
					Network:      "mainnet",
					From:         "tz1gHJt7J1aEtW2wpCR5RJd3CpnbVxUTaEXS",
					To:           "tz1gsJENNUwg7fQiRwQi5zJYaj7YtwwsE3y2",
					AmountBigInt: big.NewInt(1000000000),
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			params := types.NewParameters([]byte(tt.operation.Parameters))
			node := new(base.Node)
			if err := json.Unmarshal(params.Value, node); err != nil {
				t.Errorf("Unmarshal() error = %v", err)
				return
			}

			if err := tt.tree.ParseValue(node); err != nil {
				t.Errorf("ParseValue() error = %v", err)
				return
			}
			got, err := MakeFa2Transfers(tt.tree, tt.operation)
			if (err != nil) != tt.wantErr {
				t.Errorf("MakeFa2Transfers() error = %v, wantErr %v", err, tt.wantErr)
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
