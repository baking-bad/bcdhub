package ast

import (
	"testing"

	"github.com/baking-bad/bcdhub/internal/bcd/types"
	"github.com/stretchr/testify/require"
)

const bigMapNatNatTree = `{"prim":"big_map","args":[{"prim":"nat"},{"prim":"nat"}]}`

// bigMapDiff - diff for big_map(nat, nat). Empty `value` means the key is removed.
func bigMapDiff(ptr int64, key, value string) *types.BigMapDiff {
	diff := &types.BigMapDiff{
		Ptr: ptr,
		Key: []byte(`{"int":"` + key + `"}`),
	}
	if value != "" {
		diff.Value = []byte(`{"int":"` + value + `"}`)
	}
	return diff
}

// newEnrichedBigMap - big_map(nat, nat) with pointer 5 enriched by diffs.
func newEnrichedBigMap(t *testing.T, diffs ...*types.BigMapDiff) *TypedAst {
	t.Helper()
	typ, err := NewTypedAstFromString(bigMapNatNatTree)
	require.NoError(t, err)
	require.NoError(t, typ.SettleFromBytes([]byte(`{"int":"5"}`)))
	require.NoError(t, typ.EnrichBigMap(diffs))
	return typ
}

// key 1 is updated, key 2 is removed
func newBigMapWithRemovedKey(t *testing.T) *TypedAst {
	t.Helper()
	return newEnrichedBigMap(t,
		bigMapDiff(5, "1", "10"),
		bigMapDiff(5, "2", ""),
	)
}

func TestBigMap_RemovedKey_GetJSONModel(t *testing.T) {
	typ := newBigMapWithRemovedKey(t)

	model := make(JSONModel)
	require.NotPanics(t, func() {
		typ.GetJSONModel(model)
	})

	b, err := json.MarshalToString(model)
	require.NoError(t, err)
	require.JSONEq(t, `{"@big_map_1":[{"@nat_2":"1","@nat_3":"10"}]}`, b)
}

func TestBigMap_RemovedKey_ToParameters(t *testing.T) {
	typ := newBigMapWithRemovedKey(t)

	var (
		params []byte
		err    error
	)
	require.NotPanics(t, func() {
		params, err = typ.ToParameters("")
	})
	require.NoError(t, err)
	require.Equal(t, `[{"prim":"Elt","args":[{"int":"1"},{"int":"10"}]}]`, string(params))
}

func TestBigMap_RemovedKey_ToBaseNode(t *testing.T) {
	typ := newBigMapWithRemovedKey(t)

	node, err := typ.ToBaseNode(false)
	require.NoError(t, err)

	b, err := json.MarshalToString(node)
	require.NoError(t, err)
	require.JSONEq(t, `[{"prim":"Elt","args":[{"int":"1"},{"int":"10"}]}]`, b)
}

func TestBigMap_RemovedKey_String(t *testing.T) {
	typ := newBigMapWithRemovedKey(t)
	// Ptr is set, so String prints the pointer: reset it to print the data
	typ.Nodes[0].(*BigMap).Ptr = nil

	var s string
	require.NotPanics(t, func() {
		s = typ.String()
	})
	require.Contains(t, s, "<removed>")
}

func TestBigMap_RemovedKey_Diff(t *testing.T) {
	type child struct {
		name     string
		diffType string
		value    interface{}
	}
	tests := []struct {
		name    string
		current []*types.BigMapDiff
		prev    []*types.BigMapDiff
		want    []child
	}{
		{
			name:    "key removed in previous storage and absent in current",
			current: []*types.BigMapDiff{bigMapDiff(5, "3", "30")},
			prev: []*types.BigMapDiff{
				bigMapDiff(5, "1", "10"),
				bigMapDiff(5, "2", ""),
			},
			want: []child{
				{name: "3", diffType: MiguelKindCreate, value: "30"},
				{name: "1", diffType: MiguelKindDelete, value: "10"},
			},
		}, {
			name:    "key removed in current storage",
			current: []*types.BigMapDiff{bigMapDiff(5, "2", "")},
			prev:    []*types.BigMapDiff{bigMapDiff(5, "2", "20")},
			want: []child{
				{name: "2", diffType: MiguelKindDelete, value: "20"},
			},
		}, {
			name:    "key removed in both storages",
			current: []*types.BigMapDiff{bigMapDiff(5, "2", "")},
			prev:    []*types.BigMapDiff{bigMapDiff(5, "2", "")},
			want:    []child{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			current := newEnrichedBigMap(t, tt.current...)
			prev := newEnrichedBigMap(t, tt.prev...)

			var (
				diff *MiguelNode
				err  error
			)
			require.NotPanics(t, func() {
				diff, err = current.Diff(prev)
			})
			require.NoError(t, err)
			require.Len(t, diff.Children, len(tt.want))

			for i, want := range tt.want {
				got := diff.Children[i]
				require.NotNil(t, got.Name)
				require.Equal(t, want.name, *got.Name)
				require.Equal(t, want.diffType, got.DiffType)
				require.Equal(t, want.value, got.Value)
			}
		})
	}
}
