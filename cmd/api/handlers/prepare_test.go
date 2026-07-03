package handlers

import (
	"context"
	"testing"

	"github.com/baking-bad/bcdhub/internal/cache"
	"github.com/baking-bad/bcdhub/internal/config"
	mock_protocol "github.com/baking-bad/bcdhub/internal/models/mock/protocol"
	"github.com/baking-bad/bcdhub/internal/models/operation"
	"github.com/baking-bad/bcdhub/internal/models/protocol"
	modelTypes "github.com/baking-bad/bcdhub/internal/models/types"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

// Regression test: events emitted without payload (e.g. `EMIT ... unit`)
// must not fail with a JSON unmarshal error on empty bytes.
func TestPrepareOperations_Event(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	protocols := mock_protocol.NewMockRepository(ctrl)
	protocols.EXPECT().
		GetByID(gomock.Any(), int64(1)).
		Return(protocol.Protocol{
			ID:   1,
			Hash: "PtSeouLouXkxhg39oWzjxDWaCydNfR3RxCUrNe4Q9Ro8BTehcbh",
		}, nil).
		AnyTimes()

	cfgCtx := &config.Context{
		Network: modelTypes.Shadownet,
		Cache:   cache.NewCache(nil, nil, nil, protocols, nil),
	}

	t.Run("without payload", func(t *testing.T) {
		ops := []operation.Operation{
			{
				Kind:        modelTypes.OperationKindEvent,
				ProtocolID:  1,
				PayloadType: []byte(`{"prim":"unit"}`),
			},
		}

		resp, err := PrepareOperations(context.Background(), cfgCtx, ops, false)
		require.NoError(t, err)
		require.Len(t, resp, 1)
		require.Nil(t, resp[0].Payload)
	})

	t.Run("with payload", func(t *testing.T) {
		ops := []operation.Operation{
			{
				Kind:        modelTypes.OperationKindEvent,
				ProtocolID:  1,
				PayloadType: []byte(`{"prim":"nat"}`),
				Payload:     []byte(`{"int":"42"}`),
			},
		}

		resp, err := PrepareOperations(context.Background(), cfgCtx, ops, false)
		require.NoError(t, err)
		require.Len(t, resp, 1)
		require.Len(t, resp[0].Payload, 1)
		require.Equal(t, "42", resp[0].Payload[0].Value)
	})
}
