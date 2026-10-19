package ast

import (
	"testing"

	"github.com/baking-bad/bcdhub/internal/bcd/consts"
	"github.com/stretchr/testify/require"
)

func TestOption_ParseValue(t *testing.T) {
	tests := []struct {
		name       string
		tree       string
		data       string
		wantValue  string
		wantParams string
		wantErr    error
	}{
		{
			name:       "None",
			tree:       `{"prim":"option","args":[{"prim":"nat"}]}`,
			data:       `{"prim":"None"}`,
			wantValue:  consts.None,
			wantParams: `{"prim":"None"}`,
		}, {
			name:       "Some",
			tree:       `{"prim":"option","args":[{"prim":"nat"}]}`,
			data:       `{"prim":"Some","args":[{"int":"5"}]}`,
			wantValue:  consts.Some,
			wantParams: `{"prim":"Some","args":[{"int":"5"}]}`,
		}, {
			name:       "Some with pair",
			tree:       `{"prim":"option","args":[{"prim":"pair","args":[{"prim":"nat"},{"prim":"bool"}]}]}`,
			data:       `{"prim":"Some","args":[{"prim":"Pair","args":[{"int":"5"},{"prim":"True"}]}]}`,
			wantValue:  consts.Some,
			wantParams: `{"prim":"Some","args":[{"prim":"Pair","args":[{"int":"5"},{"prim":"True"}]}]}`,
		}, {
			name:    "Some without args",
			tree:    `{"prim":"option","args":[{"prim":"nat"}]}`,
			data:    `{"prim":"Some"}`,
			wantErr: consts.ErrInvalidArgsCount,
		}, {
			name:    "Some with empty args",
			tree:    `{"prim":"option","args":[{"prim":"key"}]}`,
			data:    `{"prim":"Some","args":[]}`,
			wantErr: consts.ErrInvalidArgsCount,
		}, {
			name:    "Some with two args",
			tree:    `{"prim":"option","args":[{"prim":"nat"}]}`,
			data:    `{"prim":"Some","args":[{"int":"1"},{"int":"2"}]}`,
			wantErr: consts.ErrInvalidArgsCount,
		}, {
			name:    "None with args",
			tree:    `{"prim":"option","args":[{"prim":"nat"}]}`,
			data:    `{"prim":"None","args":[{"int":"1"}]}`,
			wantErr: consts.ErrInvalidArgsCount,
		}, {
			name:    "unknown prim",
			tree:    `{"prim":"option","args":[{"prim":"nat"}]}`,
			data:    `{"prim":"Left","args":[{"int":"1"}]}`,
			wantErr: consts.ErrInvalidPrim,
		}, {
			name:    "invalid inner value",
			tree:    `{"prim":"option","args":[{"prim":"bool"}]}`,
			data:    `{"prim":"Some","args":[{"int":"1"}]}`,
			wantErr: consts.ErrInvalidPrim,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			typ, err := NewTypedAstFromString(tt.tree)
			require.NoError(t, err)

			require.NotPanics(t, func() {
				err = typ.SettleFromBytes([]byte(tt.data))
			})
			if tt.wantErr != nil {
				require.ErrorIs(t, err, tt.wantErr)
				require.Nil(t, typ.Nodes[0].GetValue(), "option must stay unset on error")
				return
			}
			require.NoError(t, err)
			require.Equal(t, tt.wantValue, typ.Nodes[0].GetValue())

			params, err := typ.ToParameters("")
			require.NoError(t, err)
			require.Equal(t, tt.wantParams, string(params))
		})
	}
}

func TestOption_ParseValueInvalidSomeKeepsPreviousValue(t *testing.T) {
	typ, err := NewTypedAstFromString(`{"prim":"option","args":[{"prim":"key"}]}`)
	require.NoError(t, err)
	require.NoError(t, typ.SettleFromBytes([]byte(`{"prim":"None"}`)))

	err = typ.SettleFromBytes([]byte(`{"prim":"Some"}`))
	require.ErrorIs(t, err, consts.ErrInvalidArgsCount)
	require.Equal(t, consts.None, typ.Nodes[0].GetValue())

	// the tree must stay usable: previously Some with unset key panicked here
	require.NotPanics(t, func() {
		_, err = typ.ToMiguel()
	})
	require.NoError(t, err)
}

const namedOptionTree = `{"prim":"option","args":[{"prim":"nat","annots":["%n"]}],"annots":["%o"]}`

func TestOption_FromJSONSchema(t *testing.T) {
	tests := []struct {
		name       string
		tree       string
		data       map[string]interface{}
		wantValue  string
		wantParams string
		wantErr    error
	}{
		{
			name:       "Some",
			tree:       namedOptionTree,
			data:       map[string]interface{}{"o": map[string]interface{}{"schemaKey": "Some", "n": 5.0}},
			wantValue:  consts.Some,
			wantParams: `{"prim":"Some","args":[{"int":"5"}]}`,
		}, {
			name:       "None",
			tree:       namedOptionTree,
			data:       map[string]interface{}{"o": map[string]interface{}{"schemaKey": "None"}},
			wantValue:  consts.None,
			wantParams: `{"prim":"None"}`,
		}, {
			name:       "top-level option without own key",
			tree:       `{"prim":"option","args":[{"prim":"nat"}]}`,
			data:       map[string]interface{}{"schemaKey": "Some", "@nat_2": 7.0},
			wantValue:  consts.Some,
			wantParams: `{"prim":"Some","args":[{"int":"7"}]}`,
		}, {
			name: "own key has priority over parent's schemaKey",
			tree: namedOptionTree,
			data: map[string]interface{}{
				"schemaKey": "",
				"o":         map[string]interface{}{"schemaKey": "Some", "n": 5.0},
			},
			wantValue:  consts.Some,
			wantParams: `{"prim":"Some","args":[{"int":"5"}]}`,
		}, {
			name:    "missing schemaKey",
			tree:    namedOptionTree,
			data:    map[string]interface{}{"o": map[string]interface{}{"n": 5.0}},
			wantErr: consts.ErrJSONDataIsAbsent,
		}, {
			name:    "missing option",
			tree:    namedOptionTree,
			data:    map[string]interface{}{"another": 1.0},
			wantErr: consts.ErrJSONDataIsAbsent,
		}, {
			name:    "null option",
			tree:    namedOptionTree,
			data:    map[string]interface{}{"o": nil},
			wantErr: consts.ErrValidation,
		}, {
			name:    "option is not an object",
			tree:    namedOptionTree,
			data:    map[string]interface{}{"o": "Some"},
			wantErr: consts.ErrValidation,
		}, {
			name:    "schemaKey is not a string",
			tree:    namedOptionTree,
			data:    map[string]interface{}{"o": map[string]interface{}{"schemaKey": 1.0}},
			wantErr: consts.ErrValidation,
		}, {
			name:    "unknown schemaKey",
			tree:    namedOptionTree,
			data:    map[string]interface{}{"o": map[string]interface{}{"schemaKey": "Maybe"}},
			wantErr: consts.ErrValidation,
		}, {
			name:    "invalid inner value",
			tree:    namedOptionTree,
			data:    map[string]interface{}{"o": map[string]interface{}{"schemaKey": "Some", "n": "abc"}},
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
				require.Nil(t, typ.Nodes[0].GetValue(), "option must stay unset on error")
				return
			}
			require.NoError(t, err)
			require.Equal(t, tt.wantValue, typ.Nodes[0].GetValue())

			params, err := typ.ToParameters("")
			require.NoError(t, err)
			require.Equal(t, tt.wantParams, string(params))
		})
	}
}

func TestOption_FromJSONSchemaDoesNotModifyInput(t *testing.T) {
	for _, data := range []map[string]interface{}{
		// nested under the option's own name
		{"o": map[string]interface{}{"schemaKey": "Some", "n": 5.0}},
		// top-level: no own key, the option reads `data` itself
		{"schemaKey": "Some", "n": 5.0},
	} {
		typ, err := NewTypedAstFromString(namedOptionTree)
		require.NoError(t, err)

		before, err := json.MarshalToString(data)
		require.NoError(t, err)

		require.NoError(t, typ.FromJSONSchema(data))

		after, err := json.MarshalToString(data)
		require.NoError(t, err)
		require.JSONEq(t, before, after)

		// the same input can be reused
		require.NoError(t, typ.FromJSONSchema(data))
		params, err := typ.ToParameters("")
		require.NoError(t, err)
		require.Equal(t, `{"prim":"Some","args":[{"int":"5"}]}`, string(params))
	}
}

func TestOption_FromJSONSchemaSome(t *testing.T) {
	typ, err := NewTypedAstFromString(namedOptionTree)
	require.NoError(t, err)
	require.NoError(t, typ.FromJSONSchema(map[string]interface{}{
		"o": map[string]interface{}{"schemaKey": "Some", "n": 5.0},
	}))

	opt := typ.Nodes[0].(*Option)
	require.True(t, opt.IsSome())

	model := make(JSONModel)
	typ.GetJSONModel(model)
	b, err := json.MarshalToString(model)
	require.NoError(t, err)
	require.JSONEq(t, `{"o":{"schemaKey":"Some","n":"5"}}`, b)
}
