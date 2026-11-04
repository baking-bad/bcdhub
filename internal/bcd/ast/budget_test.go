package ast

import (
	"strings"
	"testing"
	"time"

	"github.com/baking-bad/bcdhub/internal/bcd/consts"
	"github.com/stretchr/testify/require"
)

func pairOfUnits(n int) string {
	return `{"prim":"pair","args":[` + strings.TrimSuffix(strings.Repeat(`{"prim":"unit"},`, n), ",") + `]}`
}

func emptyItems(n int) []any {
	items := make([]any, n)
	for i := range items {
		items[i] = map[string]any{}
	}
	return items
}

// A wide element type multiplied by many tiny `{}` items used to take
// seconds of CPU and gigabytes of memory before failing or succeeding.
func TestFromJSONSchema_InstantiationBudget(t *testing.T) {
	typ, err := NewTypedAstFromString(`{"prim":"list","args":[` + pairOfUnits(2000) + `]}`)
	require.NoError(t, err)

	start := time.Now()
	err = typ.FromJSONSchema(map[string]any{"@list_1": emptyItems(4000)})
	require.ErrorIs(t, err, consts.ErrValidation)
	require.Less(t, time.Since(start), perfTestLimit)
}

func TestFromJSONSchema_InstantiationBudget_Nested(t *testing.T) {
	typ, err := NewTypedAstFromString(`{"prim":"list","args":[{"prim":"list","args":[` + pairOfUnits(2000) + `]}]}`)
	require.NoError(t, err)

	outer := make([]any, 2000)
	for i := range outer {
		outer[i] = map[string]any{"@list_2": emptyItems(1)}
	}
	err = typ.FromJSONSchema(map[string]any{"@list_1": outer})
	require.ErrorIs(t, err, consts.ErrValidation)
}

func TestFromJSONSchema_InstantiationBudget_Map(t *testing.T) {
	typ, err := NewTypedAstFromString(`{"prim":"map","args":[{"prim":"nat"},` + pairOfUnits(2000) + `]}`)
	require.NoError(t, err)

	err = typ.FromJSONSchema(map[string]any{"@map_1": emptyItems(1000)})
	require.ErrorIs(t, err, consts.ErrValidation)
}

func TestFromJSONSchema_InstantiationBudget_WithinLimit(t *testing.T) {
	typ, err := NewTypedAstFromString(`{"prim":"list","args":[{"prim":"nat"}]}`)
	require.NoError(t, err)

	items := make([]any, 10000)
	for i := range items {
		items[i] = map[string]any{"@nat_2": "1"}
	}
	require.NoError(t, typ.FromJSONSchema(map[string]any{"@list_1": items}))
}

func TestTypeSize(t *testing.T) {
	typ, err := NewTypedAstFromString(`{"prim":"map","args":[{"prim":"nat"},{"prim":"pair","args":[{"prim":"list","args":[{"prim":"unit"}]},{"prim":"or","args":[{"prim":"int"},{"prim":"string"}]}]}]}`)
	require.NoError(t, err)
	require.Equal(t, 8, typeSize(typ.Nodes[0]))
}

// The budget pays for copies actually made: a large element type of an empty
// collection must not count against elements of another collection.
func TestFromJSONSchema_InstantiationBudget_CountsActualCopies(t *testing.T) {
	typ, err := NewTypedAstFromString(`{"prim":"pair","args":[` +
		`{"prim":"list","annots":["%big"],"args":[` + pairOfUnits(200) + `]},` +
		`{"prim":"list","annots":["%small"],"args":[{"prim":"nat","annots":["%v"]}]}]}`)
	require.NoError(t, err)

	items := make([]any, 10000)
	for i := range items {
		items[i] = map[string]any{"v": "1"}
	}
	require.NoError(t, typ.FromJSONSchema(map[string]any{"big": []any{}, "small": items}))
}

func TestNodeBudget_Instantiate(t *testing.T) {
	typ, err := NewTypedAstFromString(`{"prim":"list","args":[{"prim":"pair","args":[{"prim":"nat"},{"prim":"list","args":[{"prim":"unit"}]}]}]}`)
	require.NoError(t, err)
	element := typ.Nodes[0].(*List).Type
	require.Equal(t, 4, typeSize(element))

	b := &nodeBudget{left: 10}
	first, err := b.instantiate(element)
	require.NoError(t, err)
	_, err = b.instantiate(element)
	require.NoError(t, err)
	_, err = b.instantiate(element)
	require.ErrorIs(t, err, consts.ErrValidation)
	require.Equal(t, 2, b.left)

	// collections inside a copy pay from the same budget
	inner := first.(*Pair).Args[1].(*List)
	require.Same(t, b, inner.budget)

	// without a budget instantiate is a plain copy
	var none *nodeBudget
	copied, err := none.instantiate(element)
	require.NoError(t, err)
	require.Nil(t, copied.(*Pair).Args[1].(*List).budget)
}

// A wide pair is parsed as a deep right comb; serializing it used to copy
// every level's bytes into its parent, quadratic in the width.
func TestPair_ToParameters_WideCombIsLinear(t *testing.T) {
	const width = 100000
	typ, err := NewTypedAstFromString(`{"prim":"pair","args":[` + strings.TrimSuffix(strings.Repeat(`{"prim":"unit"},`, width), ",") + `]}`)
	require.NoError(t, err)

	start := time.Now()
	b, err := typ.Nodes[0].ToParameters()
	require.NoError(t, err)
	require.Less(t, time.Since(start), perfTestLimit)
	require.Equal(t, width-1, strings.Count(string(b), `"Pair"`))
}

// perfTestLimit bounds performance regression tests. It is generous on
// purpose: the linear code takes a fraction of it even on a slow runner, while
// the quadratic code these tests guard against takes tens of seconds.
const perfTestLimit = 5 * time.Second
