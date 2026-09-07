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
