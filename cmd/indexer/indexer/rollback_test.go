package indexer

import (
	"context"
	"testing"

	"github.com/baking-bad/bcdhub/internal/config"
	"github.com/baking-bad/bcdhub/internal/models/block"
	mockblock "github.com/baking-bad/bcdhub/internal/models/mock/block"
	"github.com/baking-bad/bcdhub/internal/models/types"
	"github.com/baking-bad/bcdhub/internal/noderpc"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestGetLastRollbackBlock(t *testing.T) {
	errRPC := errors.New("node is unavailable")
	errStorage := errors.New("no rows in result set")

	tests := []struct {
		name       string
		startLevel int64
		stateLevel int64
		setup      func(rpc *noderpc.MockINode, blocks *mockblock.MockRepository)
		want       int64
		wantErr    error
	}{
		{
			name:       "common ancestor one level below the head",
			startLevel: 100,
			stateLevel: 105,
			setup: func(rpc *noderpc.MockINode, blocks *mockblock.MockRepository) {
				rpc.EXPECT().GetHeader(gomock.Any(), int64(105)).
					Return(noderpc.Header{Level: 105, Predecessor: "block_104"}, nil)
				blocks.EXPECT().Get(gomock.Any(), int64(104)).
					Return(block.Block{Level: 104, Hash: "block_104"}, nil)
			},
			want: 104,
		},
		{
			name:       "walks down until predecessors match",
			startLevel: 100,
			stateLevel: 105,
			setup: func(rpc *noderpc.MockINode, blocks *mockblock.MockRepository) {
				rpc.EXPECT().GetHeader(gomock.Any(), int64(105)).
					Return(noderpc.Header{Level: 105, Predecessor: "node_104"}, nil)
				blocks.EXPECT().Get(gomock.Any(), int64(104)).
					Return(block.Block{Level: 104, Hash: "orphan_104"}, nil)

				rpc.EXPECT().GetHeader(gomock.Any(), int64(104)).
					Return(noderpc.Header{Level: 104, Predecessor: "node_103"}, nil)
				blocks.EXPECT().Get(gomock.Any(), int64(103)).
					Return(block.Block{Level: 103, Hash: "orphan_103"}, nil)

				rpc.EXPECT().GetHeader(gomock.Any(), int64(103)).
					Return(noderpc.Header{Level: 103, Predecessor: "common_102"}, nil)
				blocks.EXPECT().Get(gomock.Any(), int64(102)).
					Return(block.Block{Level: 102, Hash: "common_102"}, nil)
			},
			want: 102,
		},
		{
			// startLevel+1 must still be probed: it is the lowest level where
			// Blocks.Get(level-1) can resolve against the locally indexed history.
			name:       "common ancestor exactly at the start level",
			startLevel: 100,
			stateLevel: 101,
			setup: func(rpc *noderpc.MockINode, blocks *mockblock.MockRepository) {
				rpc.EXPECT().GetHeader(gomock.Any(), int64(101)).
					Return(noderpc.Header{Level: 101, Predecessor: "block_100"}, nil)
				blocks.EXPECT().Get(gomock.Any(), int64(100)).
					Return(block.Block{Level: 100, Hash: "block_100"}, nil)
			},
			want: 100,
		},
		{
			name:       "reorg deeper than the locally indexed history",
			startLevel: 100,
			stateLevel: 102,
			setup: func(rpc *noderpc.MockINode, blocks *mockblock.MockRepository) {
				rpc.EXPECT().GetHeader(gomock.Any(), int64(102)).
					Return(noderpc.Header{Level: 102, Predecessor: "node_101"}, nil)
				blocks.EXPECT().Get(gomock.Any(), int64(101)).
					Return(block.Block{Level: 101, Hash: "orphan_101"}, nil)

				rpc.EXPECT().GetHeader(gomock.Any(), int64(101)).
					Return(noderpc.Header{Level: 101, Predecessor: "node_100"}, nil)
				blocks.EXPECT().Get(gomock.Any(), int64(100)).
					Return(block.Block{Level: 100, Hash: "orphan_100"}, nil)
				// level 100 == startLevel: the walk stops instead of running past it
			},
			wantErr: errDeepReorg,
		},
		{
			name:       "state is already at the start level",
			startLevel: 100,
			stateLevel: 100,
			setup:      func(rpc *noderpc.MockINode, blocks *mockblock.MockRepository) {},
			wantErr:    errDeepReorg,
		},
		{
			name:       "rpc error is propagated",
			startLevel: 100,
			stateLevel: 105,
			setup: func(rpc *noderpc.MockINode, blocks *mockblock.MockRepository) {
				rpc.EXPECT().GetHeader(gomock.Any(), int64(105)).
					Return(noderpc.Header{}, errRPC)
			},
			wantErr: errRPC,
		},
		{
			name:       "storage error is propagated",
			startLevel: 100,
			stateLevel: 105,
			setup: func(rpc *noderpc.MockINode, blocks *mockblock.MockRepository) {
				rpc.EXPECT().GetHeader(gomock.Any(), int64(105)).
					Return(noderpc.Header{Level: 105, Predecessor: "block_104"}, nil)
				blocks.EXPECT().Get(gomock.Any(), int64(104)).
					Return(block.Block{}, errStorage)
			},
			wantErr: errStorage,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			rpc := noderpc.NewMockINode(ctrl)
			blocks := mockblock.NewMockRepository(ctrl)
			tt.setup(rpc, blocks)

			bi := &BlockchainIndexer{
				Context: &config.Context{
					RPC:    rpc,
					Blocks: blocks,
				},
				Network:    types.Mainnet,
				startLevel: tt.startLevel,
				state:      block.Block{Level: tt.stateLevel},
			}

			got, err := bi.getLastRollbackBlock(context.Background())
			if tt.wantErr != nil {
				require.ErrorIs(t, err, tt.wantErr)
				require.Zero(t, got)
				return
			}

			require.NoError(t, err)
			require.Equal(t, tt.want, got)
		})
	}
}
