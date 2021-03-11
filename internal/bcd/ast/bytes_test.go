package ast

import (
	"reflect"
	"testing"
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
				Name:  getStringPtr("@bytes_1"),
				Value: `{ Pair "ledger" "tz1b1L4P8P1ucwuqCEP1Hxs7KB68CX8prFCp" }`,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tree, err := NewSettledTypedAst(tt.tree, tt.node)
			if err != nil {
				t.Errorf("NewSettledTypedAst error %v", err)
				return
			}
			got, err := tree.Nodes[0].ToMiguel()
			if (err != nil) != tt.wantErr {
				t.Errorf("Bytes.ToMiguel() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Bytes.ToMiguel() = %v, want %v", got, tt.want)
			}
		})
	}
}

func getStringPtr(val string) *string {
	return &val
}
