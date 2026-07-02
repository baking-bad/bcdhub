package handlers

import (
	"context"
	"testing"
	"time"

	"github.com/baking-bad/bcdhub/internal/bcd/consts"
	"github.com/baking-bad/bcdhub/internal/models/protocol"
	"github.com/baking-bad/bcdhub/internal/noderpc"
	"github.com/stretchr/testify/require"
)

// Regression test: internal operations without amount and destination
// (e.g. events) must not panic on nil dereference.
func TestParseAppliedRunCode_EventOperation(t *testing.T) {
	main := &Operation{
		Network:     "shadownet",
		Protocol:    "PtSeouLouXkxhg39oWzjxDWaCydNfR3RxCUrNe4Q9Ro8BTehcbh",
		Timestamp:   time.Now().UTC(),
		Level:       1000,
		Destination: "KT1FSSJtvHsRL3orgtj8mBi4EoaKJyStG2FF",
		Status:      consts.Applied,
	}

	response := noderpc.RunCodeResponse{
		Operations: []noderpc.Operation{
			{
				Kind:   "event",
				Source: "KT1FSSJtvHsRL3orgtj8mBi4EoaKJyStG2FF",
			},
		},
	}

	operations, err := parseAppliedRunCode(context.Background(), nil, response, nil, main, protocol.Protocol{})
	require.NoError(t, err)
	require.Len(t, operations, 2)

	event := operations[1]
	require.Equal(t, "event", event.Kind)
	require.EqualValues(t, 0, event.Amount)
	require.Empty(t, event.Destination)
	require.True(t, event.Internal)
}
