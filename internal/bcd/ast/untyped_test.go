package ast

import (
	"testing"

	"github.com/stretchr/testify/require"
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
			err := json.UnmarshalFromString(tt.tree, &tree)
			require.NoError(t, err)

			got, err := tree.GetStrings(tt.tryUnpack)
			require.Equal(t, tt.wantErr, err != nil)
			if err != nil {
				return
			}
			require.Equal(t, tt.want, got)
		})
	}
}
