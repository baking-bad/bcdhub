package ast

import (
	"testing"

	"github.com/baking-bad/bcdhub/internal/bcd/consts"
	"github.com/stretchr/testify/require"
)

func TestParseType_ArgsCount(t *testing.T) {
	tests := []struct {
		name    string
		tree    string
		wantErr bool
	}{
		// list
		{name: "list", tree: `{"prim":"list","args":[{"prim":"nat"}]}`},
		{name: "list without args", tree: `{"prim":"list"}`, wantErr: true},
		{name: "list with empty args", tree: `{"prim":"list","args":[]}`, wantErr: true},
		{name: "list with 2 args", tree: `{"prim":"list","args":[{"prim":"nat"},{"prim":"int"}]}`, wantErr: true},

		// set
		{name: "set", tree: `{"prim":"set","args":[{"prim":"nat"}]}`},
		{name: "set without args", tree: `{"prim":"set"}`, wantErr: true},
		{name: "set with empty args", tree: `{"prim":"set","args":[]}`, wantErr: true},
		{name: "set with 2 args", tree: `{"prim":"set","args":[{"prim":"nat"},{"prim":"int"}]}`, wantErr: true},

		// lambda
		{name: "lambda", tree: `{"prim":"lambda","args":[{"prim":"nat"},{"prim":"unit"}]}`},
		{name: "lambda without args", tree: `{"prim":"lambda"}`, wantErr: true},
		{name: "lambda with 1 arg", tree: `{"prim":"lambda","args":[{"prim":"nat"}]}`, wantErr: true},
		{name: "lambda with 3 args", tree: `{"prim":"lambda","args":[{"prim":"nat"},{"prim":"unit"},{"prim":"int"}]}`, wantErr: true},

		// option
		{name: "option", tree: `{"prim":"option","args":[{"prim":"nat"}]}`},
		{name: "option without args", tree: `{"prim":"option"}`, wantErr: true},
		{name: "option with empty args", tree: `{"prim":"option","args":[]}`, wantErr: true},
		{name: "option with 2 args", tree: `{"prim":"option","args":[{"prim":"nat"},{"prim":"int"}]}`, wantErr: true},

		// or
		{name: "or", tree: `{"prim":"or","args":[{"prim":"nat"},{"prim":"unit"}]}`},
		{name: "or without args", tree: `{"prim":"or"}`, wantErr: true},
		{name: "or with empty args", tree: `{"prim":"or","args":[]}`, wantErr: true},
		{name: "or with 1 arg", tree: `{"prim":"or","args":[{"prim":"nat"}]}`, wantErr: true},
		{name: "or with 3 args", tree: `{"prim":"or","args":[{"prim":"nat"},{"prim":"unit"},{"prim":"int"}]}`, wantErr: true},

		// invalid child deep in the tree
		{name: "or with 1 arg inside or", tree: `{"prim":"or","args":[{"prim":"or","args":[{"prim":"nat"}]},{"prim":"unit"}]}`, wantErr: true},
		{name: "list without args inside pair", tree: `{"prim":"pair","args":[{"prim":"nat"},{"prim":"list"}]}`, wantErr: true},
		{name: "set without args inside map value", tree: `{"prim":"map","args":[{"prim":"nat"},{"prim":"set"}]}`, wantErr: true},
		{name: "lambda with 1 arg inside option", tree: `{"prim":"option","args":[{"prim":"lambda","args":[{"prim":"nat"}]}]}`, wantErr: true},
		{name: "option without args inside list", tree: `{"prim":"list","args":[{"prim":"option"}]}`, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var (
				typ *TypedAst
				err error
			)
			require.NotPanics(t, func() {
				typ, err = NewTypedAstFromString(tt.tree)
			})
			if tt.wantErr {
				require.ErrorIs(t, err, consts.ErrInvalidArgsCount)
				return
			}
			require.NoError(t, err)
			require.Len(t, typ.Nodes, 1)
		})
	}
}

func TestTicket_ParseType_Comparable(t *testing.T) {
	tests := []struct {
		name    string
		tree    string
		wantErr bool
	}{
		{name: "ticket nat", tree: `{"prim":"ticket","args":[{"prim":"nat"}]}`},
		{name: "ticket pair", tree: `{"prim":"ticket","args":[{"prim":"pair","args":[{"prim":"nat"},{"prim":"option","args":[{"prim":"bytes"}]}]}]}`},
		{name: "ticket or", tree: `{"prim":"ticket","args":[{"prim":"or","args":[{"prim":"string"},{"prim":"address"}]}]}`},
		{name: "list of tickets", tree: `{"prim":"list","args":[{"prim":"ticket","args":[{"prim":"unit"}]}]}`},
		{name: "ticket ticket", tree: `{"prim":"ticket","args":[{"prim":"ticket","args":[{"prim":"nat"}]}]}`, wantErr: true},
		{name: "ticket list", tree: `{"prim":"ticket","args":[{"prim":"list","args":[{"prim":"nat"}]}]}`, wantErr: true},
		{name: "ticket big_map", tree: `{"prim":"ticket","args":[{"prim":"big_map","args":[{"prim":"nat"},{"prim":"nat"}]}]}`, wantErr: true},
		{name: "ticket in pair", tree: `{"prim":"ticket","args":[{"prim":"pair","args":[{"prim":"nat"},{"prim":"ticket","args":[{"prim":"nat"}]}]}]}`, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewTypedAstFromString(tt.tree)
			if tt.wantErr {
				require.ErrorIs(t, err, consts.ErrTypeIsNotComparable)
				return
			}
			require.NoError(t, err)
		})
	}
}

// Deeply nested tickets used to double the type tree at every level.
func TestTicket_ParseType_NestedIsRejectedFast(t *testing.T) {
	tree := `{"prim":"nat"}`
	for range 40 {
		tree = `{"prim":"ticket","args":[` + tree + `]}`
	}
	_, err := NewTypedAstFromString(tree)
	require.ErrorIs(t, err, consts.ErrTypeIsNotComparable)
}
