package indexer

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/baking-bad/bcdhub/internal/config"
	mockprotocol "github.com/baking-bad/bcdhub/internal/models/mock/protocol"
	"github.com/baking-bad/bcdhub/internal/models/protocol"
	"github.com/baking-bad/bcdhub/internal/models/types"
	"github.com/baking-bad/bcdhub/internal/noderpc"
	"github.com/baking-bad/bcdhub/internal/postgres/core"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/require"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
	"go.uber.org/mock/gomock"
)

const (
	protoP1 = "PtP1"
	protoP2 = "PtP2"
)

func TestRefreshProtocol(t *testing.T) {
	ctrl := gomock.NewController(t)
	protocols := mockprotocol.NewMockRepository(ctrl)

	p1 := protocol.Protocol{ID: 1, Hash: protoP1, StartLevel: 1, Constants: &protocol.Constants{TimeBetweenBlocks: 10}}
	protocols.EXPECT().Get(gomock.Any(), "", int64(99)).Return(p1, nil)

	bi := &BlockchainIndexer{
		Context:         &config.Context{Protocols: protocols},
		Network:         types.Mainnet,
		currentProtocol: protocol.Protocol{ID: 2, Hash: protoP2, StartLevel: 100, Constants: &protocol.Constants{TimeBetweenBlocks: 5}},
	}
	bi.blockTime.Store(int64(5 * time.Second))

	require.NoError(t, bi.refreshProtocol(context.Background(), 99))
	require.Equal(t, p1, bi.currentProtocol)
	require.EqualValues(t, 10*time.Second, bi.blockTime.Load())
}

func TestRefreshProtocolError(t *testing.T) {
	ctrl := gomock.NewController(t)
	protocols := mockprotocol.NewMockRepository(ctrl)

	errStorage := errors.New("no rows in result set")
	protocols.EXPECT().Get(gomock.Any(), "", int64(99)).Return(protocol.Protocol{}, errStorage)

	p2 := protocol.Protocol{ID: 2, Hash: protoP2, StartLevel: 100, Constants: &protocol.Constants{TimeBetweenBlocks: 5}}
	bi := &BlockchainIndexer{
		Context:         &config.Context{Protocols: protocols},
		Network:         types.Mainnet,
		currentProtocol: p2,
	}
	bi.blockTime.Store(int64(5 * time.Second))

	err := bi.refreshProtocol(context.Background(), 99)
	require.ErrorIs(t, err, errStorage)
	require.Equal(t, p2, bi.currentProtocol, "protocol must not change on error")
	require.EqualValues(t, 5*time.Second, bi.blockTime.Load())
}

// TestNeedsMigrationAfterRollback reproduces the reorg on an activation block:
// before the fix the cached P2 survived the rollback and the new activation block
// with P2 skipped the migration.
func TestNeedsMigrationAfterRollback(t *testing.T) {
	ctrl := gomock.NewController(t)
	protocols := mockprotocol.NewMockRepository(ctrl)

	p1 := protocol.Protocol{ID: 1, Hash: protoP1, StartLevel: 1, Constants: &protocol.Constants{TimeBetweenBlocks: 10}}
	protocols.EXPECT().Get(gomock.Any(), "", int64(99)).Return(p1, nil)

	bi := &BlockchainIndexer{
		Context:         &config.Context{Protocols: protocols},
		Network:         types.Mainnet,
		currentProtocol: protocol.Protocol{ID: 2, Hash: protoP2, StartLevel: 100},
	}

	activation := noderpc.Header{Level: 100, Protocol: protoP2}
	require.False(t, bi.needsMigration(activation), "with P2 cached the activation block is skipped")

	require.NoError(t, bi.refreshProtocol(context.Background(), 99))
	require.True(t, bi.needsMigration(activation), "after rollback to P1 the activation block must migrate")
}

func TestNeedsMigration(t *testing.T) {
	tests := []struct {
		name   string
		header noderpc.Header
		want   bool
	}{
		{name: "same protocol", header: noderpc.Header{Level: 100, Protocol: protoP1}, want: false},
		{name: "new protocol", header: noderpc.Header{Level: 100, Protocol: protoP2}, want: true},
		{name: "first block", header: noderpc.Header{Level: 1, Protocol: protoP1}, want: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bi := &BlockchainIndexer{currentProtocol: protocol.Protocol{Hash: protoP1}}
			require.Equal(t, tt.want, bi.needsMigration(tt.header))
		})
	}
}

// newUnreachableStorage returns storage whose transactions never start: every
// call made with a cancelled context fails before touching the network, which
// lets a test see that doMigration reached the migration step.
func newUnreachableStorage(t *testing.T) *core.Postgres {
	t.Helper()
	sqlDB, err := sql.Open("pgx", "postgres://user:password@127.0.0.1:1/db")
	require.NoError(t, err)
	db := bun.NewDB(sqlDB, pgdialect.New())
	t.Cleanup(func() { require.NoError(t, db.Close()) })
	return &core.Postgres{DB: db}
}

func TestDoMigration(t *testing.T) {
	t.Run("same protocol is skipped", func(t *testing.T) {
		bi := &BlockchainIndexer{
			Context:         &config.Context{},
			Network:         types.Mainnet,
			currentProtocol: protocol.Protocol{Hash: protoP2},
		}
		// no storage: reaching the migration would panic
		require.NoError(t, bi.doMigration(context.Background(), noderpc.Header{Level: 100, Protocol: protoP2}))
	})

	t.Run("activation block after rollback migrates", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		protocols := mockprotocol.NewMockRepository(ctrl)

		p1 := protocol.Protocol{ID: 1, Hash: protoP1, StartLevel: 1, Constants: &protocol.Constants{TimeBetweenBlocks: 10}}
		protocols.EXPECT().Get(gomock.Any(), "", int64(99)).Return(p1, nil)

		bi := &BlockchainIndexer{
			Context: &config.Context{
				Protocols: protocols,
				StorageDB: newUnreachableStorage(t),
			},
			Network:         types.Mainnet,
			currentProtocol: protocol.Protocol{ID: 2, Hash: protoP2, StartLevel: 100},
		}
		require.NoError(t, bi.refreshProtocol(context.Background(), 99))

		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		err := bi.doMigration(ctx, noderpc.Header{Level: 100, Protocol: protoP2})
		require.ErrorIs(t, err, context.Canceled)
		require.ErrorContains(t, err, "create postgres transaction", "the migration must be started")
	})
}

func TestReloadProtocolRetries(t *testing.T) {
	setProtocolReloadDelay(t, time.Millisecond)

	ctrl := gomock.NewController(t)
	protocols := mockprotocol.NewMockRepository(ctrl)

	p1 := protocol.Protocol{ID: 1, Hash: protoP1, StartLevel: 1, Constants: &protocol.Constants{TimeBetweenBlocks: 10}}
	gomock.InOrder(
		protocols.EXPECT().Get(gomock.Any(), "", int64(99)).Return(protocol.Protocol{}, errors.New("connection reset")).Times(2),
		protocols.EXPECT().Get(gomock.Any(), "", int64(99)).Return(p1, nil),
	)

	bi := &BlockchainIndexer{
		Context:         &config.Context{Protocols: protocols},
		Network:         types.Mainnet,
		currentProtocol: protocol.Protocol{ID: 2, Hash: protoP2, StartLevel: 100},
	}

	require.NoError(t, bi.reloadProtocol(context.Background(), 99))
	require.Equal(t, p1, bi.currentProtocol)
}

func TestReloadProtocolStopsOnContextDone(t *testing.T) {
	setProtocolReloadDelay(t, time.Millisecond)

	ctrl := gomock.NewController(t)
	protocols := mockprotocol.NewMockRepository(ctrl)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var calls int
	protocols.EXPECT().Get(gomock.Any(), "", int64(99)).DoAndReturn(func(context.Context, string, int64) (protocol.Protocol, error) {
		calls++
		if calls == 3 {
			cancel()
		}
		return protocol.Protocol{}, errors.New("connection reset")
	}).MinTimes(3)

	p2 := protocol.Protocol{ID: 2, Hash: protoP2, StartLevel: 100}
	bi := &BlockchainIndexer{
		Context:         &config.Context{Protocols: protocols},
		Network:         types.Mainnet,
		currentProtocol: p2,
	}

	err := bi.reloadProtocol(ctx, 99)
	require.ErrorIs(t, err, context.Canceled)
	require.Equal(t, p2, bi.currentProtocol)
}

func setProtocolReloadDelay(t *testing.T, d time.Duration) {
	t.Helper()
	previous := protocolReloadDelay
	protocolReloadDelay = d
	t.Cleanup(func() { protocolReloadDelay = previous })
}
