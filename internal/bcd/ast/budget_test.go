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
	require.Less(t, time.Since(start), time.Second)
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
	require.Equal(t, 7, maxElementTypeSize(typ.Nodes[0]))
}

// A wide pair is parsed as a deep right comb; serializing it used to copy
// every level's bytes into its parent, quadratic in the width.
func TestPair_ToParameters_WideCombIsLinear(t *testing.T) {
	const width = 40000
	typ, err := NewTypedAstFromString(`{"prim":"pair","args":[` + strings.TrimSuffix(strings.Repeat(`{"prim":"unit"},`, width), ",") + `]}`)
	require.NoError(t, err)

	start := time.Now()
	b, err := typ.Nodes[0].ToParameters()
	require.NoError(t, err)
	require.Less(t, time.Since(start), time.Second)
	require.Equal(t, width-1, strings.Count(string(b), `"Pair"`))
}
