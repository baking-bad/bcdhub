package ast

import (
	"testing"

	"github.com/baking-bad/bcdhub/internal/bcd/consts"
	"github.com/stretchr/testify/require"
)

const (
	flatOrTree   = `{"prim":"or","args":[{"prim":"nat","annots":["%a"]},{"prim":"unit","annots":["%b"]}],"annots":["%o"]}`
	nestedOrTree = `{"prim":"or","args":[{"prim":"or","args":[{"prim":"nat","annots":["%a"]},{"prim":"string","annots":["%b"]}],"annots":["%inner"]},{"prim":"unit","annots":["%c"]}],"annots":["%o"]}`
)

func TestOr_FromJSONSchema(t *testing.T) {
	tests := []struct {
		name    string
		tree    string
		data    map[string]interface{}
		want    string
		wantErr error
	}{
		{
			name: "left",
			tree: flatOrTree,
			data: map[string]interface{}{"o": map[string]interface{}{"schemaKey": "L", "a": 5.0}},
			want: `{"prim":"Left","args":[{"int":"5"}]}`,
		}, {
			name: "right",
			tree: flatOrTree,
			data: map[string]interface{}{"o": map[string]interface{}{"schemaKey": "R"}},
			want: `{"prim":"Right","args":[{"prim":"Unit"}]}`,
		}, {
			name: "nested left right",
			tree: nestedOrTree,
			data: map[string]interface{}{"o": map[string]interface{}{"schemaKey": "LR", "b": "text"}},
			want: `{"prim":"Left","args":[{"prim":"Right","args":[{"string":"text"}]}]}`,
		}, {
			name: "nested right",
			tree: nestedOrTree,
			data: map[string]interface{}{"o": map[string]interface{}{"schemaKey": "R"}},
			want: `{"prim":"Right","args":[{"prim":"Unit"}]}`,
		}, {
			name:    "empty schemaKey",
			tree:    flatOrTree,
			data:    map[string]interface{}{"o": map[string]interface{}{"schemaKey": ""}},
			wantErr: consts.ErrValidation,
		}, {
			name:    "missing schemaKey",
			tree:    flatOrTree,
			data:    map[string]interface{}{"o": map[string]interface{}{"a": 5.0}},
			wantErr: consts.ErrJSONDataIsAbsent,
		}, {
			name:    "missing or value",
			tree:    flatOrTree,
			data:    map[string]interface{}{"another": map[string]interface{}{"schemaKey": "L"}},
			wantErr: consts.ErrJSONDataIsAbsent,
		}, {
			name:    "schemaKey is not a string",
			tree:    flatOrTree,
			data:    map[string]interface{}{"o": map[string]interface{}{"schemaKey": 1.0}},
			wantErr: consts.ErrValidation,
		}, {
			name:    "or value is not an object",
			tree:    flatOrTree,
			data:    map[string]interface{}{"o": "L"},
			wantErr: consts.ErrValidation,
		}, {
			name:    "invalid schemaKey char",
			tree:    flatOrTree,
			data:    map[string]interface{}{"o": map[string]interface{}{"schemaKey": "X"}},
			wantErr: consts.ErrValidation,
		}, {
			name:    "schemaKey too long for flat or",
			tree:    flatOrTree,
			data:    map[string]interface{}{"o": map[string]interface{}{"schemaKey": "LL", "a": 5.0}},
			wantErr: consts.ErrValidation,
		}, {
			name:    "schemaKey too short for nested or",
			tree:    nestedOrTree,
			data:    map[string]interface{}{"o": map[string]interface{}{"schemaKey": "L", "a": 5.0}},
			wantErr: consts.ErrValidation,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			typ, err := NewTypedAstFromString(tt.tree)
			require.NoError(t, err)

			require.NotPanics(t, func() {
				err = typ.FromJSONSchema(tt.data)
			})
			if tt.wantErr != nil {
				require.ErrorIs(t, err, tt.wantErr)
				return
			}
			require.NoError(t, err)

			params, err := typ.ToParameters("")
			require.NoError(t, err)
			require.Equal(t, tt.want, string(params))
		})
	}
}

func TestOr_FromJSONSchemaDoesNotModifyInput(t *testing.T) {
	typ, err := NewTypedAstFromString(nestedOrTree)
	require.NoError(t, err)

	data := map[string]interface{}{"o": map[string]interface{}{"schemaKey": "LL", "a": 5.0}}
	before, err := json.MarshalToString(data)
	require.NoError(t, err)

	require.NoError(t, typ.FromJSONSchema(data))

	// previously the nested `or` got a self-referencing map, which can't be marshaled at all
	after, err := json.MarshalToString(data)
	require.NoError(t, err)
	require.JSONEq(t, before, after)

	// the same input can be reused
	require.NoError(t, typ.FromJSONSchema(data))
	params, err := typ.ToParameters("")
	require.NoError(t, err)
	require.Equal(t, `{"prim":"Left","args":[{"prim":"Left","args":[{"int":"5"}]}]}`, string(params))
}

func TestOr_FromJSONSchemaWithOption(t *testing.T) {
	// parent's schemaKey must not reach the option: otherwise the option may pick it
	// instead of its own value depending on map iteration order
	for range 100 {
		typ, err := NewTypedAstFromString(`{"prim":"or","args":[{"prim":"option","args":[{"prim":"nat"}]},{"prim":"unit"}]}`)
		require.NoError(t, err)

		params, err := typ.ParametersForExecution(consts.DefaultEntrypoint, map[string]interface{}{
			"@or_1": map[string]interface{}{
				"schemaKey": "L",
				"@option_2": map[string]interface{}{"schemaKey": "Some", "@nat_3": 5.0},
			},
		})
		require.NoError(t, err)
		require.Equal(t, `{"prim":"Left","args":[{"prim":"Some","args":[{"int":"5"}]}]}`, string(params.Value))
	}
}

// right branch is bool: unlike unit/nat it validates its value, so an invalid inner value fails
const natOrBoolTree = `{"prim":"or","args":[{"prim":"nat"},{"prim":"bool"}]}`

func TestOr_ParseValue(t *testing.T) {
	tests := []struct {
		name       string
		data       string
		wantParams string
		wantErr    error
	}{
		{
			name:       "left",
			data:       `{"prim":"Left","args":[{"int":"5"}]}`,
			wantParams: `{"prim":"Left","args":[{"int":"5"}]}`,
		}, {
			name:       "right",
			data:       `{"prim":"Right","args":[{"prim":"True"}]}`,
			wantParams: `{"prim":"Right","args":[{"prim":"True"}]}`,
		}, {
			name:    "left without args",
			data:    `{"prim":"Left"}`,
			wantErr: consts.ErrInvalidArgsCount,
		}, {
			name:    "right with empty args",
			data:    `{"prim":"Right","args":[]}`,
			wantErr: consts.ErrInvalidArgsCount,
		}, {
			name:    "left with 2 args",
			data:    `{"prim":"Left","args":[{"int":"5"},{"int":"6"}]}`,
			wantErr: consts.ErrInvalidArgsCount,
		}, {
			name:    "unknown prim",
			data:    `{"prim":"Some","args":[{"int":"5"}]}`,
			wantErr: consts.ErrInvalidPrim,
		}, {
			name:    "invalid inner value",
			data:    `{"prim":"Right","args":[{"int":"5"}]}`,
			wantErr: consts.ErrInvalidPrim,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			typ, err := NewTypedAstFromString(natOrBoolTree)
			require.NoError(t, err)

			require.NotPanics(t, func() {
				err = typ.SettleFromBytes([]byte(tt.data))
			})
			if tt.wantErr != nil {
				require.ErrorIs(t, err, tt.wantErr)
				require.Zero(t, typ.Nodes[0].(*Or).key, "or must stay unset on error")
				return
			}
			require.NoError(t, err)

			params, err := typ.ToParameters("")
			require.NoError(t, err)
			require.Equal(t, tt.wantParams, string(params))
		})
	}
}

func TestOr_ParseValueInvalidKeepsPreviousBranch(t *testing.T) {
	typ, err := NewTypedAstFromString(natOrBoolTree)
	require.NoError(t, err)
	require.NoError(t, typ.SettleFromBytes([]byte(`{"prim":"Left","args":[{"int":"5"}]}`)))

	err = typ.SettleFromBytes([]byte(`{"prim":"Right","args":[{"int":"5"}]}`))
	require.ErrorIs(t, err, consts.ErrInvalidPrim)

	params, err := typ.ToParameters("")
	require.NoError(t, err)
	require.Equal(t, `{"prim":"Left","args":[{"int":"5"}]}`, string(params))
}
