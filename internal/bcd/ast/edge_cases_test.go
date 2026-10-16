package ast

import (
	"testing"

	"github.com/baking-bad/bcdhub/internal/bcd/base"
	"github.com/baking-bad/bcdhub/internal/bcd/consts"
	"github.com/baking-bad/bcdhub/internal/bcd/types"
	"github.com/stretchr/testify/require"
)

func TestTypedAst_EmptyTree(t *testing.T) {
	typ, err := NewTypedAstFromString(`[]`)
	require.NoError(t, err)
	require.Empty(t, typ.Nodes)

	require.NotPanics(t, func() {
		require.Nil(t, typ.FindByName(consts.DefaultEntrypoint, true))
		require.Nil(t, typ.FindByName("entrypoint_0", true))
		require.Nil(t, typ.FindByName("a", false))

		params, err := typ.ToParameters(consts.DefaultEntrypoint)
		require.NoError(t, err)
		require.Nil(t, params)

		docs, err := typ.Docs(consts.DefaultEntrypoint)
		require.NoError(t, err)
		require.Nil(t, docs)
	})
}

func TestGetAnnotation_EmptyAnnotation(t *testing.T) {
	var typ *TypedAst
	var err error
	require.NotPanics(t, func() {
		typ, err = NewTypedAstFromString(`{"prim":"nat","annots":["","%amount",""]}`)
	})
	require.NoError(t, err)
	require.Equal(t, "amount", typ.Nodes[0].GetName())

	require.Equal(t, "", getAnnotation([]string{""}, consts.AnnotPrefixFieldName))
	require.Equal(t, "", getAnnotation([]string{"%"}, consts.AnnotPrefixFieldName))
	require.Equal(t, "t", getAnnotation([]string{"", ":t"}, consts.AnnotPrefixrefixTypeName))
}

func TestDefault_DistinguishMixedValues(t *testing.T) {
	str := func(s string) *string { return &s }
	tests := []struct {
		name      string
		current   interface{}
		prev      interface{}
		wantDiff  string
		wantValue interface{}
		wantFrom  interface{}
	}{
		{
			name:      "int and string",
			current:   types.NewBigInt(1),
			prev:      "1",
			wantDiff:  MiguelKindUpdate,
			wantValue: "1",
			wantFrom:  "1",
		}, {
			name:      "string and int",
			current:   "1",
			prev:      types.NewBigInt(1),
			wantDiff:  MiguelKindUpdate,
			wantValue: "1",
			wantFrom:  "1",
		}, {
			name:      "int and empty int",
			current:   types.NewBigInt(1),
			prev:      &types.BigInt{},
			wantDiff:  MiguelKindUpdate,
			wantValue: "1",
			wantFrom:  "<nil>",
		}, {
			name:      "empty ints are equal",
			current:   &types.BigInt{},
			prev:      &types.BigInt{},
			wantValue: "<nil>",
		}, {
			name:      "equal ints",
			current:   types.NewBigInt(5),
			prev:      types.NewBigInt(5),
			wantValue: "5",
		}, {
			name:      "uncomparable values",
			current:   map[string]interface{}{"a": 1.0},
			prev:      map[string]interface{}{"a": 2.0},
			wantDiff:  MiguelKindUpdate,
			wantValue: map[string]interface{}{"a": 1.0},
			wantFrom:  map[string]interface{}{"a": 2.0},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			current := NewNat(0)
			current.Value = tt.current
			prev := NewNat(0)
			prev.Value = tt.prev

			var (
				node *MiguelNode
				err  error
			)
			require.NotPanics(t, func() {
				node, err = current.Distinguish(prev)
			})
			require.NoError(t, err)
			require.Equal(t, tt.wantDiff, node.DiffType)
			require.Equal(t, tt.wantValue, node.Value)
			require.Equal(t, tt.wantFrom, node.From)

			require.NotPanics(t, func() {
				current.Equal(prev)
			})
		})
	}

	// a value parsed from a string literal for nat (see п.23) must not break the diff
	parsed := NewNat(0)
	require.NoError(t, parsed.ParseValue(&base.Node{StringValue: str("1")}))
	require.NotPanics(t, func() {
		_, err := newSettledNat(t, 1).Distinguish(parsed)
		require.NoError(t, err)
	})
}

// newSettledNat - settled nat for tests
func newSettledNat(t *testing.T, value int64) *Nat {
	t.Helper()
	nat := NewNat(0)
	require.NoError(t, nat.ParseValue(&base.Node{IntValue: types.NewBigInt(value)}))
	return nat
}

func TestDistinguish_ContainersWithIncomparableChildren(t *testing.T) {
	tests := []struct {
		name         string
		currentTree  string
		currentValue string
		prevTree     string
		prevValue    string
		wantChildren int
	}{
		{
			name:         "pair with option and unsettled previous value",
			currentTree:  `{"prim":"pair","args":[{"prim":"option","args":[{"prim":"nat"}]},{"prim":"nat"}]}`,
			currentValue: `{"prim":"Pair","args":[{"prim":"None"},{"int":"1"}]}`,
			prevTree:     `{"prim":"pair","args":[{"prim":"option","args":[{"prim":"nat"}]},{"prim":"nat"}]}`,
			wantChildren: 2,
		}, {
			name:         "pair with different arg types",
			currentTree:  `{"prim":"pair","args":[{"prim":"nat"},{"prim":"nat"}]}`,
			currentValue: `{"prim":"Pair","args":[{"int":"2"},{"int":"1"}]}`,
			prevTree:     `{"prim":"pair","args":[{"prim":"string"},{"prim":"nat"}]}`,
			prevValue:    `{"prim":"Pair","args":[{"string":"x"},{"int":"1"}]}`,
			wantChildren: 2,
		}, {
			name:         "or with different branch types",
			currentTree:  `{"prim":"or","args":[{"prim":"nat"},{"prim":"unit"}]}`,
			currentValue: `{"prim":"Left","args":[{"int":"1"}]}`,
			prevTree:     `{"prim":"or","args":[{"prim":"string"},{"prim":"unit"}]}`,
			prevValue:    `{"prim":"Left","args":[{"string":"x"}]}`,
			wantChildren: 1,
		}, {
			name:         "option with different types",
			currentTree:  `{"prim":"option","args":[{"prim":"nat"}]}`,
			currentValue: `{"prim":"Some","args":[{"int":"1"}]}`,
			prevTree:     `{"prim":"option","args":[{"prim":"string"}]}`,
			prevValue:    `{"prim":"Some","args":[{"string":"x"}]}`,
		}, {
			name:         "map with different value types",
			currentTree:  `{"prim":"map","args":[{"prim":"nat"},{"prim":"nat"}]}`,
			currentValue: `[{"prim":"Elt","args":[{"int":"1"},{"int":"1"}]}]`,
			prevTree:     `{"prim":"map","args":[{"prim":"nat"},{"prim":"string"}]}`,
			prevValue:    `[{"prim":"Elt","args":[{"int":"1"},{"string":"x"}]}]`,
			wantChildren: 1,
		}, {
			name:         "list with different item types",
			currentTree:  `{"prim":"list","args":[{"prim":"nat"}]}`,
			currentValue: `[{"int":"1"}]`,
			prevTree:     `{"prim":"list","args":[{"prim":"string"}]}`,
			prevValue:    `[{"string":"x"}]`,
			wantChildren: 1,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			current, err := NewTypedAstFromString(tt.currentTree)
			require.NoError(t, err)
			require.NoError(t, current.SettleFromBytes([]byte(tt.currentValue)))

			prev, err := NewTypedAstFromString(tt.prevTree)
			require.NoError(t, err)
			if tt.prevValue != "" {
				require.NoError(t, prev.SettleFromBytes([]byte(tt.prevValue)))
			}

			var diff *MiguelNode
			require.NotPanics(t, func() {
				diff, err = current.Diff(prev)
			})
			require.NoError(t, err)
			require.NotNil(t, diff)
			require.Len(t, diff.Children, tt.wantChildren)
			for _, child := range diff.Children {
				require.NotNil(t, child)
			}
		})
	}
}

func TestPair_DistinguishDifferentArgsCount(t *testing.T) {
	current := NewPair(0)
	current.Args = []Node{newSettledNat(t, 1), newSettledNat(t, 2)}
	prev := NewPair(0)
	prev.Args = []Node{newSettledNat(t, 1)}

	var err error
	require.NotPanics(t, func() {
		_, err = current.Distinguish(prev)
	})
	require.ErrorIs(t, err, consts.ErrTreesAreDifferent)
}
