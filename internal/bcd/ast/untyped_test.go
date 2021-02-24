package ast

import (
	"reflect"
	"testing"
)

func TestUntypedAST_GetStrings(t *testing.T) {
	tests := []struct {
		name      string
		tree      string
		want      []string
		tryUnpack bool
		wantErr   bool
	}{
		{
			name:      "test 1",
			tree:      `{"bytes":"74657a6f732d73746f726167653a6d65746164617461"}`,
			tryUnpack: true,
			want:      []string{"tezos-storage:metadata"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var tree UntypedAST
			if err := json.UnmarshalFromString(tt.tree, &tree); err != nil {
				t.Errorf("UntypedAST.GetStrings() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			got, err := tree.GetStrings(tt.tryUnpack)
			if (err != nil) != tt.wantErr {
				t.Errorf("UntypedAST.GetStrings() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("UntypedAST.GetStrings() = %v, want %v", got, tt.want)
			}
		})
	}
}
