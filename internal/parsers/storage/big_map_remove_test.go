package storage

import (
	"context"
	"testing"
	"time"

	"github.com/baking-bad/bcdhub/internal/models/bigmapdiff"
	mock_bmd "github.com/baking-bad/bcdhub/internal/models/mock/bigmapdiff"
	"github.com/baking-bad/bcdhub/internal/models/operation"
	"github.com/baking-bad/bcdhub/internal/models/types"
	"github.com/baking-bad/bcdhub/internal/noderpc"
	"github.com/baking-bad/bcdhub/internal/parsers"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

// Removing a whole big map must stamp every state it touches with the level
// and time of the removal. Otherwise the upsert keeps the old
// last_update_level, and rolling back the removal block would not find these
// states, leaving the keys removed.
func TestHandleBigMapDiffRemove_StampsStatesWithRemovalLevel(t *testing.T) {
	const (
		address = "KT1Pz65ssbPF7Zv9Dh7ggqUkgAYNSuJ9iia7"
		ptr     = int64(42)
	)
	oldTime := time.Date(2022, 1, 25, 16, 45, 9, 0, time.UTC)
	opTime := time.Date(2022, 1, 25, 17, 17, 47, 0, time.UTC)

	parsersUnderTest := []struct {
		name   string
		remove func(repo bigmapdiff.Repository, op *operation.Operation, store parsers.Store) error
	}{
		{
			name: "babylon",
			remove: func(repo bigmapdiff.Repository, op *operation.Operation, store parsers.Store) error {
				b := &Babylon{bigmapdiffs: repo}
				p := ptr
				return b.handleBigMapDiffRemove(context.Background(), noderpc.BigMapDiff{Action: "remove", BigMap: &p}, address, op, store)
			},
		},
		{
			name: "lazy babylon",
			remove: func(repo bigmapdiff.Repository, op *operation.Operation, store parsers.Store) error {
				b := NewLazyBabylon(repo, nil, nil)
				return b.handleBigMapDiffRemove(context.Background(), &noderpc.LazyBigMapDiff{Action: "remove"}, ptr, address, op, store)
			},
		},
	}

	for _, tt := range parsersUnderTest {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			repo := mock_bmd.NewMockRepository(ctrl)
			repo.EXPECT().
				GetByPtr(gomock.Any(), address, ptr).
				Return([]bigmapdiff.BigMapState{
					{
						ID:              1,
						Ptr:             ptr,
						Contract:        address,
						KeyHash:         "key_hash_1",
						Key:             types.MustNewBytes("01"),
						Value:           types.MustNewBytes("0a"),
						LastUpdateLevel: 33,
						LastUpdateTime:  oldTime,
					}, {
						ID:              2,
						Ptr:             ptr,
						Contract:        address,
						KeyHash:         "key_hash_2",
						Key:             types.MustNewBytes("02"),
						Value:           types.MustNewBytes("0b"),
						LastUpdateLevel: 35,
						LastUpdateTime:  oldTime,
					},
				}, nil).
				Times(1)

			op := &operation.Operation{
				ID:         100,
				Level:      47,
				Timestamp:  opTime,
				ProtocolID: 3,
			}
			store := parsers.NewTestStore()

			require.NoError(t, tt.remove(repo, op, store))

			require.Len(t, store.BigMapState, 2)
			for _, state := range store.BigMapState {
				require.True(t, state.Removed, state.KeyHash)
				require.EqualValues(t, 47, state.LastUpdateLevel, state.KeyHash)
				require.Equal(t, opTime, state.LastUpdateTime, state.KeyHash)
			}

			require.Len(t, op.BigMapDiffs, 2)
			for _, diff := range op.BigMapDiffs {
				require.EqualValues(t, 47, diff.Level, diff.KeyHash)
				require.Equal(t, opTime, diff.Timestamp, diff.KeyHash)
				require.EqualValues(t, 100, diff.OperationID, diff.KeyHash)
				require.Nil(t, diff.Value, diff.KeyHash)
			}
		})
	}
}
