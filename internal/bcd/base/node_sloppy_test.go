//go:build jsoniter_sloppy

package base

import (
	"testing"

	"github.com/stretchr/testify/require"
)

// The indexer is built with jsoniter_sloppy, whose outer decoder accepts
// close to 10000 nested Pair. Decoding the tree with jsoniter's depth-counting
// callbacks used to halve that, so the indexer stopped at about 5000.
func TestNode_UnmarshalJSON_SloppyDepth(t *testing.T) {
	var node Node
	require.NoError(t, json.UnmarshalFromString(nestedPairs(9990), &node))
}
