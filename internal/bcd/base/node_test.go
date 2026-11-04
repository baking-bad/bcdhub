package base

import (
	"strings"
	"testing"
	"time"

	"github.com/baking-bad/bcdhub/internal/bcd/consts"
	"github.com/stretchr/testify/require"
)

func TestNode_GetAnnotations(t *testing.T) {
	tests := []struct {
		name string
		tree string
		want map[string]struct{}
	}{
		{
			name: "mainnet/KT1R3uoZ6W1ZxEwzqtv75Ro7DhVY6UAcxuK2/parameter",
			tree: `{"prim":"or","args":[{"prim":"lambda","args":[{"prim":"unit"},{"prim":"list","args":[{"prim":"operation"}]}],"annots":["%do"]},{"prim":"or","args":[{"prim":"pair","args":[{"prim":"address"},{"prim":"pair","args":[{"prim":"string"},{"prim":"option","args":[{"prim":"bytes"}]}]}]},{"prim":"or","args":[{"prim":"mutez"},{"prim":"or","args":[{"prim":"pair","args":[{"prim":"address"},{"prim":"bool"}]},{"prim":"lambda","args":[{"prim":"pair","args":[{"prim":"address"},{"prim":"pair","args":[{"prim":"string"},{"prim":"option","args":[{"prim":"bytes"}]}]}]},{"prim":"operation"}]}]}]}],"annots":["%default"]}]}`,
			want: map[string]struct{}{
				"do":      {},
				"default": {},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var node Node
			err := json.UnmarshalFromString(tt.tree, &node)
			require.NoError(t, err)
			got := node.GetAnnotations()
			require.Equal(t, tt.want, got)
		})
	}
}

func TestNode_IsComparable(t *testing.T) {
	tests := []struct {
		name string
		tree string
		want bool
	}{
		{name: "nat", tree: `{"prim":"nat"}`, want: true},
		{name: "never", tree: `{"prim":"never"}`, want: true},
		{name: "key", tree: `{"prim":"key"}`, want: true},
		{name: "signature", tree: `{"prim":"signature"}`, want: true},
		{name: "chain_id", tree: `{"prim":"chain_id"}`, want: true},
		{name: "tx_rollup_l2_address", tree: `{"prim":"tx_rollup_l2_address"}`, want: true},
		{name: "constant", tree: `{"prim":"constant","args":[{"string":"expru5X1yxJG6ezR2uHMotwMLNmSzQyh5t1vUnhjx4cS6Pv9qE1Sdo"}]}`, want: true},
		{name: "pair of comparables", tree: `{"prim":"pair","args":[{"prim":"nat"},{"prim":"address"},{"prim":"string"}]}`, want: true},
		{name: "or of comparables", tree: `{"prim":"or","args":[{"prim":"nat"},{"prim":"option","args":[{"prim":"bytes"}]}]}`, want: true},
		{name: "pair without args", tree: `{"prim":"pair"}`},
		{name: "ticket", tree: `{"prim":"ticket","args":[{"prim":"nat"}]}`},
		{name: "list", tree: `{"prim":"list","args":[{"prim":"nat"}]}`},
		{name: "set", tree: `{"prim":"set","args":[{"prim":"nat"}]}`},
		{name: "map", tree: `{"prim":"map","args":[{"prim":"nat"},{"prim":"nat"}]}`},
		{name: "big_map", tree: `{"prim":"big_map","args":[{"prim":"nat"},{"prim":"nat"}]}`},
		{name: "lambda", tree: `{"prim":"lambda","args":[{"prim":"nat"},{"prim":"nat"}]}`},
		{name: "contract", tree: `{"prim":"contract","args":[{"prim":"unit"}]}`},
		{name: "operation", tree: `{"prim":"operation"}`},
		{name: "chest", tree: `{"prim":"chest"}`},
		{name: "chest_key", tree: `{"prim":"chest_key"}`},
		{name: "bls12_381_fr", tree: `{"prim":"bls12_381_fr"}`},
		{name: "sapling_state", tree: `{"prim":"sapling_state","args":[{"int":"8"}]}`},
		{name: "baker_hash", tree: `{"prim":"baker_hash"}`},
		{name: "option of ticket", tree: `{"prim":"option","args":[{"prim":"ticket","args":[{"prim":"nat"}]}]}`},
		{name: "pair with list deep inside", tree: `{"prim":"pair","args":[{"prim":"nat"},{"prim":"or","args":[{"prim":"unit"},{"prim":"list","args":[{"prim":"nat"}]}]}]}`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var node Node
			require.NoError(t, json.UnmarshalFromString(tt.tree, &node))
			require.Equal(t, tt.want, node.IsComparable())
		})
	}

	var nilNode *Node
	require.False(t, nilNode.IsComparable())
}

func TestNode_UnmarshalJSON_RoundTrip(t *testing.T) {
	tests := []string{
		`{"prim":"Pair","args":[{"int":"-42"},{"string":"hello"}]}`,
		`{"prim":"pair","args":[{"prim":"nat","annots":["%a"]},{"prim":"bytes","annots":[":b","%c"]}]}`,
		`[{"bytes":"0a0b"},[],[{"prim":"Unit"}]]`,
		`[{"prim":"Elt","args":[{"string":"key"},null]}]`,
		`{"int":"123456789012345678901234567890"}`,
		`[]`,
	}
	for _, tt := range tests {
		t.Run(tt, func(t *testing.T) {
			var node Node
			require.NoError(t, json.UnmarshalFromString(tt, &node))
			got, err := json.MarshalToString(&node)
			require.NoError(t, err)
			require.JSONEq(t, tt, got)
		})
	}
}

func TestNode_UnmarshalJSON_Invalid(t *testing.T) {
	for _, tt := range []string{
		`"string"`,
		`42`,
		`null`,
		`{"prim":"Pair","args":{"int":"1"}}`,
		`{"prim":"Pair","args":[42]}`,
		`{"int":"abc"}`,
		`{"prim":"Pair","args":[{"int":"1"}`,
	} {
		t.Run(tt, func(t *testing.T) {
			var node Node
			require.Error(t, json.UnmarshalFromString(tt, &node))
		})
	}
}

func TestNode_UnmarshalJSON_TrailingData(t *testing.T) {
	for _, input := range []string{
		`{"prim":"Unit"}{"prim":"Other"}`,
		`[{"prim":"Unit"}] 1`,
		`{"prim":"Unit"}x`,
		`{"prim":"Unit"}}`,
		`[{"prim":"Unit"}]]`,
		`[{"prim":"Unit"}]garbage`,
		`{"prim":"Unit"} {`,
		`{"":1}`,
	} {
		var node Node
		require.Error(t, node.UnmarshalJSON([]byte(input)), input)
	}

	var node Node
	require.NoError(t, node.UnmarshalJSON([]byte(" {\"prim\":\"Unit\"} \n")))
	require.NoError(t, node.UnmarshalJSON([]byte(`[]`)))
	require.NoError(t, node.UnmarshalJSON([]byte(`{}`)))
}

func nestedPairs(n int) string {
	return strings.Repeat(`{"prim":"Pair","args":[{"int":"1"},`, n) + `{"int":"1"}` + strings.Repeat(`]}`, n)
}

// MaxNodeDepth bounds recursion when UnmarshalJSON is called directly; through
// json.Unmarshal the outer decoder limits JSON depth first.
func TestNode_UnmarshalJSON_MaxNodeDepth(t *testing.T) {
	var node Node
	require.NoError(t, node.UnmarshalJSON([]byte(nestedPairs(MaxNodeDepth-1))))
	require.ErrorIs(t, node.UnmarshalJSON([]byte(nestedPairs(MaxNodeDepth))), consts.ErrInvalidJSON)

	deepArrays := strings.Repeat("[", MaxNodeDepth+1) + strings.Repeat("]", MaxNodeDepth+1)
	require.ErrorIs(t, node.UnmarshalJSON([]byte(deepArrays)), consts.ErrInvalidJSON)
}

// The decoder must not add a depth limit of its own below the one of the
// outer decoder: every build accepts what the API accepted before.
func TestNode_UnmarshalJSON_Depth(t *testing.T) {
	var node Node
	require.NoError(t, json.UnmarshalFromString(nestedPairs(4900), &node))
}

func TestNode_UnmarshalJSON_NullResetsFields(t *testing.T) {
	var node Node
	require.NoError(t, json.UnmarshalFromString(`{"int":"1","int":null}`, &node))
	require.Nil(t, node.IntValue)

	value := "x"
	node = Node{Prim: "Pair", StringValue: &value, Args: []*Node{{Prim: "Unit"}}, Annots: []string{"%a"}}
	require.NoError(t, node.UnmarshalJSON([]byte(`{"prim":null,"args":null,"annots":null,"string":null}`)))
	require.Equal(t, Node{Prim: "Pair"}, node)
}

func TestNode_UnmarshalJSON_NullFieldsAreAbsent(t *testing.T) {
	var node Node
	require.NoError(t, json.UnmarshalFromString(`{"prim":"Unit","args":null,"annots":null,"int":null,"string":null,"bytes":null}`, &node))
	require.Equal(t, Node{Prim: "Unit"}, node)
}

// Each nesting level used to re-decode its whole subtree, so a deep tree with
// a large leaf was quadratic: 9000 levels around ~1MB took seconds.
func TestNode_UnmarshalJSON_DeepNestingIsLinear(t *testing.T) {
	const depth = 9000
	payload := `[` + strings.Repeat(`[`, depth) + `{"string":"` + strings.Repeat("a", 1<<20) + `"}` + strings.Repeat(`]`, depth) + `]`

	start := time.Now()
	var node Node
	require.NoError(t, json.UnmarshalFromString(payload, &node))
	require.Less(t, time.Since(start), perfTestLimit)
}

// perfTestLimit bounds performance regression tests. It is generous on
// purpose: the linear code takes a fraction of it even on a slow runner, while
// the quadratic code these tests guard against takes tens of seconds.
const perfTestLimit = 5 * time.Second
