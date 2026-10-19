package migrations

import (
	"testing"

	"github.com/baking-bad/bcdhub/internal/bcd/ast"
	"github.com/stretchr/testify/require"
)

func newScript(t *testing.T, parameter string) *ast.Script {
	t.Helper()
	script, err := ast.NewScriptWithoutCode([]byte(`[{"prim":"parameter","args":[` + parameter + `]},{"prim":"storage","args":[{"prim":"unit"}]},{"prim":"code","args":[[]]}]`))
	require.NoError(t, err)
	return script
}

func TestEntrypointByParameters(t *testing.T) {
	// unbalanced tree: the old names were @entrypoint_0, @entrypoint_1, @entrypoint_1, @entrypoint_1
	const unbalanced = `{"prim":"or","args":[{"prim":"or","args":[{"prim":"or","args":[{"prim":"nat"},{"prim":"int"}]},{"prim":"string"}]},{"prim":"bytes"}]}`

	tests := []struct {
		name       string
		parameter  string
		parameters string
		want       string
		wantErr    bool
	}{
		{
			name:       "first leaf",
			parameter:  unbalanced,
			parameters: `{"entrypoint":"default","value":{"prim":"Left","args":[{"prim":"Left","args":[{"prim":"Left","args":[{"int":"1"}]}]}]}}`,
			want:       "entrypoint_0",
		}, {
			name:       "second leaf",
			parameter:  unbalanced,
			parameters: `{"entrypoint":"default","value":{"prim":"Left","args":[{"prim":"Left","args":[{"prim":"Right","args":[{"int":"1"}]}]}]}}`,
			want:       "entrypoint_1",
		}, {
			name:       "third leaf",
			parameter:  unbalanced,
			parameters: `{"entrypoint":"default","value":{"prim":"Left","args":[{"prim":"Right","args":[{"string":"x"}]}]}}`,
			want:       "entrypoint_2",
		}, {
			name:       "fourth leaf",
			parameter:  unbalanced,
			parameters: `{"entrypoint":"default","value":{"prim":"Right","args":[{"bytes":"00"}]}}`,
			want:       "entrypoint_3",
		}, {
			name:       "named and unnamed leaves",
			parameter:  `{"prim":"or","args":[{"prim":"nat","annots":["%mint"]},{"prim":"or","args":[{"prim":"int"},{"prim":"string","annots":["%burn"]}]}]}`,
			parameters: `{"entrypoint":"default","value":{"prim":"Right","args":[{"prim":"Left","args":[{"int":"1"}]}]}}`,
			want:       "entrypoint_1",
		}, {
			name:       "named leaf",
			parameter:  `{"prim":"or","args":[{"prim":"nat","annots":["%mint"]},{"prim":"int"}]}`,
			parameters: `{"entrypoint":"default","value":{"prim":"Left","args":[{"int":"1"}]}}`,
			want:       "mint",
		}, {
			name:       "unnamed root",
			parameter:  `{"prim":"nat"}`,
			parameters: `{"entrypoint":"default","value":{"int":"1"}}`,
			want:       "default",
		}, {
			name:       "parameters do not match the type",
			parameter:  unbalanced,
			parameters: `{"entrypoint":"default","value":{"prim":"Right","args":[{"int":"1"},{"int":"2"}]}}`,
			wantErr:    true,
		}, {
			name:      "empty parameters",
			parameter: unbalanced,
			wantErr:   true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := entrypointByParameters(newScript(t, tt.parameter), []byte(tt.parameters))
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			require.Equal(t, tt.want, got)
		})
	}
}
