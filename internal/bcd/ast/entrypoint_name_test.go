package ast

import (
	"testing"

	"github.com/baking-bad/bcdhub/internal/bcd/consts"
	"github.com/stretchr/testify/require"
)

func TestTypedAst_UnwrapAndGetEntrypointName(t *testing.T) {
	tests := []struct {
		name string
		tree string
		// a value per leaf of the root `or`, in leaf order
		values    []string
		wantNames []string
	}{
		{
			name: "left comb",
			tree: `{"prim":"or","args":[{"prim":"or","args":[{"prim":"or","args":[{"prim":"nat"},{"prim":"int"}]},{"prim":"string"}]},{"prim":"bytes"}]}`,
			values: []string{
				`{"prim":"Left","args":[{"prim":"Left","args":[{"prim":"Left","args":[{"int":"1"}]}]}]}`,
				`{"prim":"Left","args":[{"prim":"Left","args":[{"prim":"Right","args":[{"int":"1"}]}]}]}`,
				`{"prim":"Left","args":[{"prim":"Right","args":[{"string":"x"}]}]}`,
				`{"prim":"Right","args":[{"bytes":"00"}]}`,
			},
			wantNames: []string{"entrypoint_0", "entrypoint_1", "entrypoint_2", "entrypoint_3"},
		}, {
			name: "right comb",
			tree: `{"prim":"or","args":[{"prim":"nat"},{"prim":"or","args":[{"prim":"int"},{"prim":"or","args":[{"prim":"string"},{"prim":"bytes"}]}]}]}`,
			values: []string{
				`{"prim":"Left","args":[{"int":"1"}]}`,
				`{"prim":"Right","args":[{"prim":"Left","args":[{"int":"1"}]}]}`,
				`{"prim":"Right","args":[{"prim":"Right","args":[{"prim":"Left","args":[{"string":"x"}]}]}]}`,
				`{"prim":"Right","args":[{"prim":"Right","args":[{"prim":"Right","args":[{"bytes":"00"}]}]}]}`,
			},
			wantNames: []string{"entrypoint_0", "entrypoint_1", "entrypoint_2", "entrypoint_3"},
		}, {
			name: "balanced",
			tree: `{"prim":"or","args":[{"prim":"or","args":[{"prim":"nat"},{"prim":"int"}]},{"prim":"or","args":[{"prim":"string"},{"prim":"bytes"}]}]}`,
			values: []string{
				`{"prim":"Left","args":[{"prim":"Left","args":[{"int":"1"}]}]}`,
				`{"prim":"Left","args":[{"prim":"Right","args":[{"int":"1"}]}]}`,
				`{"prim":"Right","args":[{"prim":"Left","args":[{"string":"x"}]}]}`,
				`{"prim":"Right","args":[{"prim":"Right","args":[{"bytes":"00"}]}]}`,
			},
			wantNames: []string{"entrypoint_0", "entrypoint_1", "entrypoint_2", "entrypoint_3"},
		}, {
			name: "named and unnamed",
			tree: `{"prim":"or","args":[{"prim":"nat","annots":["%a"]},{"prim":"or","args":[{"prim":"int"},{"prim":"string","annots":["%c"]}]}]}`,
			values: []string{
				`{"prim":"Left","args":[{"int":"1"}]}`,
				`{"prim":"Right","args":[{"prim":"Left","args":[{"int":"1"}]}]}`,
				`{"prim":"Right","args":[{"prim":"Right","args":[{"string":"x"}]}]}`,
			},
			wantNames: []string{"a", "entrypoint_1", "c"},
		}, {
			name: "unnamed pair entrypoint inside named or",
			tree: `{"prim":"or","args":[{"prim":"or","args":[{"prim":"unit","annots":["%x"]},{"prim":"pair","args":[{"prim":"nat"},{"prim":"nat"}]}],"annots":["%inner"]},{"prim":"unit"}]}`,
			values: []string{
				`{"prim":"Left","args":[{"prim":"Left","args":[{"prim":"Unit"}]}]}`,
				`{"prim":"Left","args":[{"prim":"Right","args":[{"prim":"Pair","args":[{"int":"1"},{"int":"2"}]}]}]}`,
				`{"prim":"Right","args":[{"prim":"Unit"}]}`,
			},
			wantNames: []string{"x", "entrypoint_1", "entrypoint_2"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			typ, err := NewTypedAstFromString(tt.tree)
			require.NoError(t, err)
			require.Equal(t, tt.wantNames, typ.GetEntrypoints())

			docs, err := typ.GetEntrypointsDocs()
			require.NoError(t, err)
			require.Len(t, docs, len(tt.wantNames))
			for i := range docs {
				require.Equal(t, tt.wantNames[i], docs[i].Name)
			}

			require.Len(t, tt.values, len(tt.wantNames))
			for i, value := range tt.values {
				settled, err := NewTypedAstFromString(tt.tree)
				require.NoError(t, err)
				require.NoError(t, settled.SettleFromBytes([]byte(value)))

				leaf, name := settled.UnwrapAndGetEntrypointName()
				require.NotNil(t, leaf)
				require.Equal(t, tt.wantNames[i], name, "value %s", value)

				// the name must point to the same leaf
				found := typ.FindByName(name, true)
				require.NotNil(t, found, "entrypoint %s is not found", name)
				// node IDs are the same for trees parsed from the same type, so names identify the leaf
				require.Equal(t, leaf.GetName(), found.GetName(), "entrypoint %s points to another leaf", name)
			}
		})
	}
}

func TestTypedAst_UnwrapAndGetEntrypointNameNotOr(t *testing.T) {
	tests := []struct {
		name string
		tree string
		data string
		want string
	}{
		{
			name: "unnamed",
			tree: `{"prim":"nat"}`,
			data: `{"int":"1"}`,
			want: consts.DefaultEntrypoint,
		}, {
			name: "named",
			tree: `{"prim":"nat","annots":["%amount"]}`,
			data: `{"int":"1"}`,
			want: "amount",
		}, {
			name: "unnamed pair",
			tree: `{"prim":"pair","args":[{"prim":"nat"},{"prim":"nat"}]}`,
			data: `{"prim":"Pair","args":[{"int":"1"},{"int":"2"}]}`,
			want: consts.DefaultEntrypoint,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			typ, err := NewTypedAstFromString(tt.tree)
			require.NoError(t, err)
			require.Equal(t, []string{tt.want}, typ.GetEntrypoints())

			require.NoError(t, typ.SettleFromBytes([]byte(tt.data)))
			node, name := typ.UnwrapAndGetEntrypointName()
			require.NotNil(t, node)
			require.Equal(t, tt.want, name)
		})
	}
}

func TestTypedAst_UnwrapAndGetEntrypointNameUnsettled(t *testing.T) {
	typ, err := NewTypedAstFromString(`{"prim":"or","args":[{"prim":"nat"},{"prim":"int"}]}`)
	require.NoError(t, err)

	node, name := typ.UnwrapAndGetEntrypointName()
	require.Nil(t, node)
	require.Empty(t, name)
}
