package ast

import (
	"testing"

	"github.com/baking-bad/bcdhub/internal/testsuite"
	"github.com/stretchr/testify/require"
)

func TestBytes_ToMiguel(t *testing.T) {
	tests := []struct {
		name    string
		node    string
		tree    string
		want    *MiguelNode
		wantErr bool
	}{
		{
			name: "tzbtc/big_map_key",
			node: `{"bytes": "05070701000000066c65646765720a000000160000a8911e6e0a3d6ae987dc908a20eab4a58e875b5d"}`,
			tree: `{"prim":"bytes"}`,
			want: &MiguelNode{
				Prim:  "bytes",
				Type:  "bytes",
				Name:  testsuite.Ptr("@bytes_1"),
				Value: `{ Pair "ledger" "tz1b1L4P8P1ucwuqCEP1Hxs7KB68CX8prFCp" }`,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tree, err := NewSettledTypedAst(tt.tree, tt.node)
			require.NoError(t, err)

			got, err := tree.Nodes[0].ToMiguel()
			require.Equal(t, tt.wantErr, err != nil)
			if err != nil {
				return
			}
			require.Equal(t, tt.want, got)
		})
	}
}
