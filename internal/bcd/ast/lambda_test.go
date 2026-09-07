package ast

import (
	"testing"

	"github.com/baking-bad/bcdhub/internal/bcd/consts"
	"github.com/stretchr/testify/require"
)

// lambdaEntrypoint is the annotation of the lambda entrypoint used in tests
const lambdaEntrypoint = "run"

func TestLambda_FromJSONSchema(t *testing.T) {
	tests := []struct {
		name      string
		typ       string
		data      map[string]interface{}
		want      string
		wantErrIs error
	}{
		{
			name: "valid michelson code",
			typ:  `{"prim":"lambda","args":[{"prim":"unit"},{"prim":"list","args":[{"prim":"operation"}]}],"annots":["%run"]}`,
			data: map[string]interface{}{lambdaEntrypoint: "{ DROP ; NIL operation }"},
			want: `[[{"prim":"DROP"},{"prim":"NIL","args":[{"prim":"operation"}]}]]`,
		}, {
			name:      "empty code",
			typ:       `{"prim":"lambda","args":[{"prim":"unit"},{"prim":"list","args":[{"prim":"operation"}]}],"annots":["%run"]}`,
			data:      map[string]interface{}{lambdaEntrypoint: ""},
			wantErrIs: consts.ErrValidation,
		}, {
			name:      "blank code",
			typ:       `{"prim":"lambda","args":[{"prim":"unit"},{"prim":"list","args":[{"prim":"operation"}]}],"annots":["%run"]}`,
			data:      map[string]interface{}{lambdaEntrypoint: "  \n\t"},
			wantErrIs: consts.ErrValidation,
		}, {
			name:      "unparsable michelson code",
			typ:       `{"prim":"lambda","args":[{"prim":"unit"},{"prim":"list","args":[{"prim":"operation"}]}],"annots":["%run"]}`,
			data:      map[string]interface{}{lambdaEntrypoint: "{ DROP"},
			wantErrIs: consts.ErrValidation,
		}, {
			name:      "number instead of code",
			typ:       `{"prim":"lambda","args":[{"prim":"unit"},{"prim":"list","args":[{"prim":"operation"}]}],"annots":["%run"]}`,
			data:      map[string]interface{}{lambdaEntrypoint: float64(42)},
			wantErrIs: consts.ErrValidation,
		}, {
			name:      "null instead of code",
			typ:       `{"prim":"lambda","args":[{"prim":"unit"},{"prim":"list","args":[{"prim":"operation"}]}],"annots":["%run"]}`,
			data:      map[string]interface{}{lambdaEntrypoint: nil},
			wantErrIs: consts.ErrValidation,
		}, {
			name: "unknown key is skipped",
			typ:  `{"prim":"lambda","args":[{"prim":"unit"},{"prim":"list","args":[{"prim":"operation"}]}],"annots":["%run"]}`,
			data: map[string]interface{}{"another_one": ""},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			typ, err := NewTypedAstFromString(tt.typ)
			require.NoError(t, err)
			require.Len(t, typ.Nodes, 1)

			node, ok := typ.Nodes[0].(*Lambda)
			require.True(t, ok)

			err = node.FromJSONSchema(tt.data)
			if tt.wantErrIs != nil {
				require.ErrorIs(t, err, tt.wantErrIs)
				return
			}
			require.NoError(t, err)

			if tt.want == "" {
				require.Nil(t, node.Value)
				return
			}
			require.Equal(t, tt.want, node.Value)
		})
	}
}

func TestLambda_ToParameters(t *testing.T) {
	const lambdaType = `{"prim":"lambda","args":[{"prim":"unit"},{"prim":"list","args":[{"prim":"operation"}]}],"annots":["%run"]}`
	const lambdaCode = `[{"prim":"DROP"},{"prim":"NIL","args":[{"prim":"operation"}]}]`

	tests := []struct {
		name      string
		prepare   func(t *testing.T, node *Lambda)
		want      string
		wantErrIs error
	}{
		{
			name: "value filled from json schema",
			prepare: func(t *testing.T, node *Lambda) {
				require.NoError(t, node.FromJSONSchema(map[string]interface{}{lambdaEntrypoint: "{ DROP ; NIL operation }"}))
			},
			want: `[[{"prim":"DROP"},{"prim":"NIL","args":[{"prim":"operation"}]}]]`,
		}, {
			name: "value parsed from micheline",
			prepare: func(t *testing.T, node *Lambda) {
				var untyped UntypedAST
				require.NoError(t, json.UnmarshalFromString(lambdaCode, &untyped))
				require.Len(t, untyped, 1)
				require.NoError(t, node.ParseValue(untyped[0]))
			},
			want: lambdaCode,
		}, {
			name:      "value is not filled",
			prepare:   func(t *testing.T, node *Lambda) {},
			wantErrIs: consts.ErrValidation,
		}, {
			name: "value is not a string",
			prepare: func(t *testing.T, node *Lambda) {
				node.Value = map[string]interface{}{"prim": "DROP"}
			},
			wantErrIs: consts.ErrValidation,
		}, {
			name: "value is an empty string",
			prepare: func(t *testing.T, node *Lambda) {
				node.Value = ""
			},
			wantErrIs: consts.ErrValidation,
		}, {
			name: "value is a blank string",
			prepare: func(t *testing.T, node *Lambda) {
				node.Value = " \n\t"
			},
			wantErrIs: consts.ErrValidation,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			typ, err := NewTypedAstFromString(lambdaType)
			require.NoError(t, err)
			require.Len(t, typ.Nodes, 1)

			node, ok := typ.Nodes[0].(*Lambda)
			require.True(t, ok)

			tt.prepare(t, node)

			params, err := node.ToParameters()
			if tt.wantErrIs != nil {
				require.ErrorIs(t, err, tt.wantErrIs)
				require.Nil(t, params)
				return
			}
			require.NoError(t, err)
			require.JSONEq(t, tt.want, string(params))
		})
	}
}

func TestLambda_ToParameters_unfilledInPair(t *testing.T) {
	typ, err := NewTypedAstFromString(`{"prim":"pair","args":[{"prim":"nat","annots":["%counter"]},{"prim":"lambda","args":[{"prim":"unit"},{"prim":"list","args":[{"prim":"operation"}]}],"annots":["%run"]}],"annots":["%do"]}`)
	require.NoError(t, err)

	_, err = typ.ParametersForExecution("do", map[string]interface{}{
		"counter": float64(1),
	})
	require.ErrorIs(t, err, consts.ErrValidation)
}
