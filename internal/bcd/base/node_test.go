package base

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNode_GetAnnotations(t *testing.T) {
	tests := []struct {
		name string
		tree string
		want map[string]struct{}
	}{
		{
			name: "mainnet/KT1R3uoZ6W1ZxEwzqtv75Ro7DhVY6UAcxuK2/parameter",
			tree: `{"prim":"or","args":[{"prim":"lambda","args":[{"prim":"unit"},{"prim":"list","args":[{"prim":"operation"}]}],"annots":["%do"]},{"prim":"or","args":[{"prim":"pair","args":[{"prim":"address"},{"prim":"pair","args":[{"prim":"string"},{"prim":"option","args":[{"prim":"bytes"}]}]}]},{"prim":"or","args":[{"prim":"mutez"},{"prim":"or","args":[{"prim":"pair","args":[{"prim":"address"},{"prim":"bool"}]},{"prim":"lambda","args":[{"prim":"pair","args":[{"prim":"address"},{"prim":"pair","args":[{"prim":"string"},{"prim":"option","args":[{"prim":"bytes"}]}]}]},{"prim":"operation"}]}]}]}],"annots":["%default"]}]}`,
			want: map[string]struct{}{
				"do":      {},
				"default": {},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var node Node
			err := json.UnmarshalFromString(tt.tree, &node)
			require.NoError(t, err)
			got := node.GetAnnotations()
			require.Equal(t, tt.want, got)
		})
	}
}
