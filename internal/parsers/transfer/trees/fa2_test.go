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
			name: "FA2: test 1",
			tree: GetFA2Transfer(),
			operation: operation.Operation{
				Network:    modelTypes.Mainnet,
				Parameters: []byte(`{"entrypoint":"transfer","value":[{"prim":"Pair","args":[{"string":"tz1gHJt7J1aEtW2wpCR5RJd3CpnbVxUTaEXS"},[{"prim":"Pair","args":[{"string":"tz1gsJENNUwg7fQiRwQi5zJYaj7YtwwsE3y2"},{"prim":"Pair","args":[{"int":"0"},{"int":"1000000000"}]}]}]]}]}`),
			},
			want: []*transfer.Transfer{
				{
					Network: modelTypes.Mainnet,
					From:    "tz1gHJt7J1aEtW2wpCR5RJd3CpnbVxUTaEXS",
					To:      "tz1gsJENNUwg7fQiRwQi5zJYaj7YtwwsE3y2",
					Amount:  newDecimal("1000000000"),
				},
			},
		}, {
			name: "FA2: test 2",
			tree: GetFA2Transfer(),
			operation: operation.Operation{
				Network:    modelTypes.Mainnet,
				Parameters: []byte(`{"entrypoint":"transfer","value":[{"prim":"Pair","args":[{"string":"tz1aCzsYRUgDZBV7zb7Si6q2AobrocFW5qwb"},[{"prim":"Pair","args":[{"string":"tz1a6ZKyEoCmfpsY74jEq6uKBK8RQXdj1aVi"},{"prim":"Pair","args":[{"int":"12"},{"int":"1"}]}]}]]}]}`),
			},
			want: []*transfer.Transfer{
				{
					Network: modelTypes.Mainnet,
					From:    "tz1aCzsYRUgDZBV7zb7Si6q2AobrocFW5qwb",
					To:      "tz1a6ZKyEoCmfpsY74jEq6uKBK8RQXdj1aVi",
					Amount:  newDecimal("1"),
					TokenID: 12,
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
