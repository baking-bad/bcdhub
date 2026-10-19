package ast

import (
	"strings"
	"testing"

	"github.com/baking-bad/bcdhub/internal/bcd/base"
	"github.com/baking-bad/bcdhub/internal/bcd/consts"
	"github.com/baking-bad/bcdhub/internal/bcd/types"
	"github.com/stretchr/testify/require"
)

func newSectionType(t *testing.T, args ...string) *SectionType {
	t.Helper()
	node := &base.Node{Prim: consts.PARAMETER}
	for _, arg := range args {
		node.Args = append(node.Args, &base.Node{Prim: arg})
	}
	st := NewSectionType(consts.PARAMETER, 0)
	var id int
	require.NoError(t, st.ParseType(node, &id))
	return st
}

func TestSectionType_ParseValue(t *testing.T) {
	st := newSectionType(t, consts.NAT, consts.STRING)

	value := "x"
	err := st.ParseValue(&base.Node{
		Prim: consts.PARAMETER,
		Args: []*base.Node{
			{IntValue: types.NewBigInt(1)},
			{StringValue: &value},
		},
	})
	require.NoError(t, err)

	// previously every value was written to the first arg
	nat, ok := st.Args[0].GetValue().(*types.BigInt)
	require.True(t, ok, "got %T", st.Args[0].GetValue())
	require.EqualValues(t, 1, nat.Int64())
	require.Equal(t, "x", st.Args[1].GetValue())
}

func TestSectionType_ParseValueArgsCount(t *testing.T) {
	st := newSectionType(t, consts.NAT, consts.STRING)

	err := st.ParseValue(&base.Node{
		Prim: consts.PARAMETER,
		Args: []*base.Node{{IntValue: types.NewBigInt(1)}},
	})
	require.ErrorIs(t, err, consts.ErrTreesAreDifferent)
}

func TestSectionType_EqualType(t *testing.T) {
	newParameter := func(args ...string) *Parameter {
		return &Parameter{SectionType: newSectionType(t, args...)}
	}

	tests := []struct {
		name string
		x    Node
		y    Node
		want bool
	}{
		{
			name: "section types",
			x:    newSectionType(t, consts.NAT),
			y:    newSectionType(t, consts.NAT),
			want: true,
		}, {
			name: "parameters",
			x:    newParameter(consts.NAT),
			y:    newParameter(consts.NAT),
			want: true,
		}, {
			name: "section type and parameter",
			x:    newSectionType(t, consts.NAT),
			y:    newParameter(consts.NAT),
			want: true,
		}, {
			name: "different args",
			x:    newSectionType(t, consts.NAT),
			y:    newSectionType(t, consts.STRING),
			want: false,
		}, {
			name: "different args count",
			x:    newSectionType(t, consts.NAT),
			y:    newSectionType(t, consts.NAT, consts.NAT),
			want: false,
		}, {
			name: "not a section",
			x:    newSectionType(t, consts.NAT),
			y:    NewNat(0),
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.Equal(t, tt.want, tt.x.EqualType(tt.y))
		})
	}
}

func TestUnit_Distinguish(t *testing.T) {
	current, err := NewTypedAstFromString(`{"prim":"unit","annots":["%current"]}`)
	require.NoError(t, err)
	require.NoError(t, current.SettleFromBytes([]byte(`{"prim":"Unit"}`)))
	prev, err := NewTypedAstFromString(`{"prim":"unit","annots":["%prev"]}`)
	require.NoError(t, err)
	require.NoError(t, prev.SettleFromBytes([]byte(`{"prim":"Unit"}`)))

	diff, err := current.Diff(prev)
	require.NoError(t, err)
	require.NotNil(t, diff.Name)
	// previously the second node was compared with itself and its name was returned
	require.Equal(t, "current", *diff.Name)
	require.Empty(t, diff.DiffType)
}

func TestLambda_FromJSONSchemaKey(t *testing.T) {
	const tree = `{"prim":"lambda","args":[{"prim":"nat"},{"prim":"nat"}],"annots":[":code","%run"]}`

	typ, err := NewTypedAstFromString(tree)
	require.NoError(t, err)

	schema, err := typ.ToJSONSchema()
	require.NoError(t, err)
	require.Contains(t, schema.Properties, "run", "schema uses the field name")

	// data comes with the key from the schema
	require.NoError(t, typ.FromJSONSchema(map[string]interface{}{"run": "{}"}))
	params, err := typ.ToParameters("")
	require.NoError(t, err)
	require.NotEmpty(t, params)

	model := make(JSONModel)
	typ.GetJSONModel(model)
	require.Contains(t, model, "run", "model uses the same key")

	// the type name is not a key of the form
	other, err := NewTypedAstFromString(tree)
	require.NoError(t, err)
	require.NoError(t, other.FromJSONSchema(map[string]interface{}{"code": "{}"}))
	_, err = other.ToParameters("")
	require.ErrorIs(t, err, consts.ErrValidation)
}

func TestTypedAst_GetJSONModelNil(t *testing.T) {
	typ, err := NewTypedAstFromString(`{"prim":"pair","args":[{"prim":"nat","annots":["%a"]},{"prim":"nat","annots":["%b"]}]}`)
	require.NoError(t, err)
	require.NotPanics(t, func() {
		typ.GetJSONModel(nil)
	})
}

func TestTypedAst_ParametersForExecutionUnnamedEntrypointErrors(t *testing.T) {
	const tree = `{"prim":"or","args":[{"prim":"nat"},{"prim":"lambda","args":[{"prim":"nat"},{"prim":"nat"}]}]}`

	tests := []struct {
		name       string
		entrypoint string
		data       map[string]interface{}
		want       string
		wantErr    error
	}{
		{
			name:       "valid",
			entrypoint: "entrypoint_1",
			data:       map[string]interface{}{"@lambda_3": "{}"},
			want:       `{"prim":"Right","args":[`,
		}, {
			name:       "invalid michelson",
			entrypoint: "entrypoint_1",
			data:       map[string]interface{}{"@lambda_3": "{ NOT_AN_INSTRUCTION"},
			wantErr:    consts.ErrValidation,
		}, {
			name:       "wrong value type",
			entrypoint: "entrypoint_0",
			data:       map[string]interface{}{"@nat_2": true},
			wantErr:    consts.ErrValidation,
		}, {
			name:       "unknown entrypoint",
			entrypoint: "entrypoint_5",
			data:       map[string]interface{}{},
			wantErr:    consts.ErrEmptyTree,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			typ, err := NewTypedAstFromString(tree)
			require.NoError(t, err)

			params, err := typ.ParametersForExecution(tt.entrypoint, tt.data)
			if tt.wantErr != nil {
				require.ErrorIs(t, err, tt.wantErr)
				require.Nil(t, params)
				return
			}
			require.NoError(t, err)
			require.Equal(t, consts.DefaultEntrypoint, params.Entrypoint)
			require.True(t, strings.HasPrefix(string(params.Value), tt.want), "got %s", params.Value)
		})
	}
}
