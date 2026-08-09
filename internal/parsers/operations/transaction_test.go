package operations

import (
	"testing"

	"github.com/baking-bad/bcdhub/internal/bcd/consts"
	"github.com/baking-bad/bcdhub/internal/models/operation"
	modelsTypes "github.com/baking-bad/bcdhub/internal/models/types"
	"github.com/stretchr/testify/require"
)

func TestTransaction_getEntrypoint(t *testing.T) {
	p := Transaction{}

	t.Run("nil AST on applied operation does not panic", func(t *testing.T) {
		tx := &operation.Operation{
			Status:     modelsTypes.OperationStatusApplied,
			Parameters: []byte(`{"entrypoint":"transfer","value":{}}`),
			AST:        nil,
		}

		require.NotPanics(t, func() {
			require.NoError(t, p.getEntrypoint(tx))
		})
		require.Equal(t, "transfer", tx.Entrypoint.Str)
	})

	t.Run("empty parameters set default entrypoint without touching AST", func(t *testing.T) {
		tx := &operation.Operation{
			Status: modelsTypes.OperationStatusApplied,
			AST:    nil,
		}

		require.NoError(t, p.getEntrypoint(tx))
		require.Equal(t, consts.DefaultEntrypoint, tx.Entrypoint.Str)
	})

	t.Run("not applied operation skips AST lookup", func(t *testing.T) {
		tx := &operation.Operation{
			Status:     modelsTypes.OperationStatusFailed,
			Parameters: []byte(`{"entrypoint":"transfer","value":{}}`),
			AST:        nil,
		}

		require.NoError(t, p.getEntrypoint(tx))
		require.Equal(t, "transfer", tx.Entrypoint.Str)
	})
}
