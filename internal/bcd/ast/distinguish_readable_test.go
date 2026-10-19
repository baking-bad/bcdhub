package ast

import (
	"strings"
	"testing"

	"github.com/baking-bad/bcdhub/internal/bcd/forge"
	"github.com/stretchr/testify/require"
)

// types whose value is stored either in optimized (bytes) or in readable (string) form
var readableTypes = []struct {
	name    string
	tree    string
	bytes   string
	changed string
	unforge func(string) (string, error)
}{
	{
		name:    "address",
		tree:    `{"prim":"address"}`,
		bytes:   "0000e7670f32038107a59a2b9cfefae36ea21f5aa63c",
		changed: "0000e7670f32038107a59a2b9cfefae36ea21f5aa63d",
		unforge: forge.UnforgeContract,
	}, {
		name:    "contract",
		tree:    `{"prim":"contract","args":[{"prim":"unit"}]}`,
		bytes:   "0000e7670f32038107a59a2b9cfefae36ea21f5aa63c",
		changed: "0000e7670f32038107a59a2b9cfefae36ea21f5aa63d",
		unforge: forge.UnforgeContract,
	}, {
		name:    "key",
		tree:    `{"prim":"key"}`,
		bytes:   "00" + strings.Repeat("ab", 32),
		changed: "00" + strings.Repeat("cd", 32),
		unforge: forge.UnforgePublicKey,
	}, {
		name:    "key_hash",
		tree:    `{"prim":"key_hash"}`,
		bytes:   "00e7670f32038107a59a2b9cfefae36ea21f5aa63c",
		changed: "00e7670f32038107a59a2b9cfefae36ea21f5aa63d",
		unforge: forge.UnforgeAddress,
	}, {
		name:    "signature",
		tree:    `{"prim":"signature"}`,
		bytes:   strings.Repeat("ab", 64),
		changed: strings.Repeat("cd", 64),
		unforge: forge.UnforgeSignature,
	}, {
		name:    "chain_id",
		tree:    `{"prim":"chain_id"}`,
		bytes:   "7a06a770",
		changed: "7a06a771",
		unforge: forge.UnforgeChainID,
	}, {
		name:    "baker_hash",
		tree:    `{"prim":"baker_hash"}`,
		bytes:   strings.Repeat("ab", 20),
		changed: strings.Repeat("cd", 20),
		unforge: forge.UnforgeBakerHash,
	}, {
		name:    "tx_rollup_l2_address",
		tree:    `{"prim":"tx_rollup_l2_address"}`,
		bytes:   strings.Repeat("ab", 20),
		changed: strings.Repeat("cd", 20),
		unforge: forge.UnforgeTxRollupL2Address,
	},
}

func settledReadableType(t *testing.T, tree, data string) *TypedAst {
	t.Helper()
	typ, err := NewTypedAstFromString(tree)
	require.NoError(t, err)
	require.NoError(t, typ.SettleFromBytes([]byte(data)))
	return typ
}

func bytesValue(hex string) string    { return `{"bytes":"` + hex + `"}` }
func stringValue(value string) string { return `{"string":"` + value + `"}` }

func TestDistinguishReadable_SameValueInDifferentForms(t *testing.T) {
	for _, tt := range readableTypes {
		t.Run(tt.name, func(t *testing.T) {
			readable, err := tt.unforge(tt.bytes)
			require.NoError(t, err)

			current := settledReadableType(t, tt.tree, bytesValue(tt.bytes))
			prev := settledReadableType(t, tt.tree, stringValue(readable))

			diff, err := current.Diff(prev)
			require.NoError(t, err)
			require.Empty(t, diff.DiffType, "the same value in bytes and string forms is not an update")
			require.Nil(t, diff.From)
			require.Equal(t, readable, diff.Value)
		})
	}
}

func TestDistinguishReadable_Update(t *testing.T) {
	for _, tt := range readableTypes {
		t.Run(tt.name, func(t *testing.T) {
			readable, err := tt.unforge(tt.bytes)
			require.NoError(t, err)
			readableChanged, err := tt.unforge(tt.changed)
			require.NoError(t, err)

			current := settledReadableType(t, tt.tree, bytesValue(tt.bytes))
			prev := settledReadableType(t, tt.tree, bytesValue(tt.changed))

			diff, err := current.Diff(prev)
			require.NoError(t, err)
			require.Equal(t, MiguelKindUpdate, diff.DiffType)
			require.Equal(t, readable, diff.Value)
			require.Equal(t, readableChanged, diff.From)
		})
	}
}

func TestDistinguishReadable_DoesNotModifyNodes(t *testing.T) {
	for _, tt := range readableTypes {
		t.Run(tt.name, func(t *testing.T) {
			readable, err := tt.unforge(tt.bytes)
			require.NoError(t, err)

			current := settledReadableType(t, tt.tree, bytesValue(tt.bytes))
			prev := settledReadableType(t, tt.tree, bytesValue(tt.changed))

			_, err = current.Diff(prev)
			require.NoError(t, err)

			// previously Distinguish replaced the hex value with the readable one in place
			require.Equal(t, tt.bytes, current.Nodes[0].GetValue())
			require.Equal(t, tt.changed, prev.Nodes[0].GetValue())

			miguel, err := current.ToMiguel()
			require.NoError(t, err)
			require.Equal(t, readable, miguel[0].Value)

			// the same trees can be compared again
			_, err = current.Diff(prev)
			require.NoError(t, err)
		})
	}
}

func TestDistinguishReadable_KnownValues(t *testing.T) {
	address := settledReadableType(t, `{"prim":"address"}`, bytesValue("0000e7670f32038107a59a2b9cfefae36ea21f5aa63c"))
	prev := settledReadableType(t, `{"prim":"address"}`, stringValue("tz1gjaF81ZRRvdzjobyfVNsAeSC6PScjfQwN"))
	diff, err := address.Diff(prev)
	require.NoError(t, err)
	require.Empty(t, diff.DiffType)
	require.Equal(t, "tz1gjaF81ZRRvdzjobyfVNsAeSC6PScjfQwN", diff.Value)

	chainID := settledReadableType(t, `{"prim":"chain_id"}`, bytesValue("7a06a770"))
	prevChainID := settledReadableType(t, `{"prim":"chain_id"}`, stringValue("NetXdQprcVkpaWU"))
	diff, err = chainID.Diff(prevChainID)
	require.NoError(t, err)
	require.Empty(t, diff.DiffType)
	require.Equal(t, "NetXdQprcVkpaWU", diff.Value)

	l2 := settledReadableType(t, `{"prim":"tx_rollup_l2_address"}`, bytesValue(strings.Repeat("ab", 20)))
	miguel, err := l2.ToMiguel()
	require.NoError(t, err)
	value, ok := miguel[0].Value.(string)
	require.True(t, ok)
	require.True(t, strings.HasPrefix(value, "tz4"), "got %s", value)
	require.Len(t, value, 36)
}
