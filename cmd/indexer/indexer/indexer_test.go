package indexer

import (
	"context"
	"net/url"
	"testing"
	"time"

	"github.com/baking-bad/bcdhub/internal/config"
	"github.com/baking-bad/bcdhub/internal/models/block"
	"github.com/baking-bad/bcdhub/internal/models/protocol"
	"github.com/baking-bad/bcdhub/internal/models/types"
	"github.com/baking-bad/bcdhub/internal/noderpc"
	"github.com/dipdup-io/workerpool"
	"github.com/getsentry/sentry-go"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

type capturingTransport struct {
	events []*sentry.Event
}

func (t *capturingTransport) Configure(options sentry.ClientOptions) {}
func (t *capturingTransport) SendEvent(event *sentry.Event) {
	t.events = append(t.events, event)
}
func (t *capturingTransport) Flush(timeout time.Duration) bool { return true }

func TestReportProcessError(t *testing.T) {
	tests := []struct {
		name string
		err  error
		sent int
	}{
		{
			name: "context canceled on shutdown",
			err: &url.Error{
				Op:  "Get",
				URL: "https://rpc.tzkt.io/mainnet/chains/main/blocks/head/header",
				Err: context.Canceled,
			},
			sent: 0,
		},
		{
			name: "transient network error",
			err:  errors.Wrap(context.DeadlineExceeded, "request head"),
			sent: 0,
		},
		{
			name: "regular error",
			err:  errors.New("unexpected response"),
			sent: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			transport := new(capturingTransport)
			client, err := sentry.NewClient(sentry.ClientOptions{Transport: transport})
			require.NoError(t, err)
			hub := sentry.NewHub(client, sentry.NewScope())

			bi := &BlockchainIndexer{
				Network: types.Mainnet,
			}
			bi.reportProcessError(hub, tt.err)

			require.Len(t, transport.events, tt.sent)
		})
	}
}

// newDeepReorgIndexer builds a BlockchainIndexer whose state is already at startLevel,
// so getLastRollbackBlock fails with errDeepReorg on its very first check, without any
// RPC or storage calls beyond the GetHead the test itself sets up. blockTime controls
// the ticker interval Start will use for its periodic re-checks.
func newDeepReorgIndexer(t *testing.T, hub *sentry.Hub, rpc *noderpc.MockINode, chainID string, blockTime time.Duration) *BlockchainIndexer {
	t.Helper()

	bi := &BlockchainIndexer{
		Context:  &config.Context{RPC: rpc},
		Network:  types.Mainnet,
		receiver: NewReceiver(rpc, 20, 2),
		blocks:   make(map[int64]*Block),
		state: block.Block{
			Level:    100,
			Protocol: protocol.Protocol{ChainID: chainID},
		},
		startLevel:   100,
		fatal:        make(chan error, 1),
		refreshTimer: make(chan time.Duration, 10),
		g:            workerpool.NewGroup(),
		hub:          hub,
	}
	bi.blockTime.Store(int64(blockTime))
	t.Cleanup(func() { require.NoError(t, bi.Close()) })
	return bi
}

// runStart runs bi.Start and waits for it to return, failing the test instead of hanging
// forever if Start does not stop on its own within the timeout.
func runStart(t *testing.T, bi *BlockchainIndexer) {
	t.Helper()

	done := make(chan struct{})
	go func() {
		defer close(done)
		bi.Start(context.Background())
	}()

	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Fatal("Start did not return: a deep reorg must stop the indexer, not spin forever")
	}
}

// TestStartStopsOnDeepReorgFirstTick pins the first of the two places identified in
// review: process() run before the ticker loop starts (the "first tick" in Start) must
// stop the indexer on errDeepReorg exactly like the ticker-driven calls do, instead of
// falling through to the loop and retrying forever.
func TestStartStopsOnDeepReorgFirstTick(t *testing.T) {
	transport := new(capturingTransport)
	client, err := sentry.NewClient(sentry.ClientOptions{Transport: transport})
	require.NoError(t, err)
	hub := sentry.NewHub(client, sentry.NewScope())

	ctrl := gomock.NewController(t)
	rpc := noderpc.NewMockINode(ctrl)
	// head.Level (50) < state.Level (100) drives process() into the Rollback branch,
	// and state.Level == startLevel makes getLastRollbackBlock fail immediately.
	rpc.EXPECT().GetHead(gomock.Any()).Return(noderpc.Header{Level: 50, ChainID: "chain-a"}, nil).Times(1)

	bi := newDeepReorgIndexer(t, hub, rpc, "chain-a", 0)

	runStart(t, bi)

	require.Len(t, transport.events, 1)
}

// TestStartStopsOnDeepReorgLoopTick pins the second place identified in review: the
// process() call made from inside the ticker loop must also stop the indexer on
// errDeepReorg. The first tick is made to fail with an unrelated, non-fatal error so it
// falls through into the loop, where the second call triggers the reorg.
func TestStartStopsOnDeepReorgLoopTick(t *testing.T) {
	transport := new(capturingTransport)
	client, err := sentry.NewClient(sentry.ClientOptions{Transport: transport})
	require.NoError(t, err)
	hub := sentry.NewHub(client, sentry.NewScope())

	ctrl := gomock.NewController(t)
	rpc := noderpc.NewMockINode(ctrl)

	var calls int
	rpc.EXPECT().GetHead(gomock.Any()).DoAndReturn(func(context.Context) (noderpc.Header, error) {
		calls++
		if calls == 1 {
			// chain-id mismatch: a regular, reported error that is not errDeepReorg,
			// so Start must survive it and enter the ticker loop.
			return noderpc.Header{Level: 150, ChainID: "wrong-chain"}, nil
		}
		// from here on, drive the same deep-reorg condition as the first-tick test.
		return noderpc.Header{Level: 50, ChainID: "chain-a"}, nil
	}).AnyTimes()

	// fast ticker so the loop's second call to process() happens quickly
	bi := newDeepReorgIndexer(t, hub, rpc, "chain-a", 20*time.Millisecond)

	runStart(t, bi)

	require.Len(t, transport.events, 2)
}

func TestTickerDuration(t *testing.T) {
	tests := []struct {
		name string
		// blockTime is the protocol's TimeBetweenBlocks in seconds, as stored by init/migrate
		blockTime int64
		seconds   int
		want      time.Duration
	}{
		{
			name:      "protocol time between blocks",
			blockTime: 8,
			seconds:   0,
			want:      8 * time.Second,
		},
		{
			name:      "explicit interval wins over protocol",
			blockTime: 8,
			seconds:   5,
			want:      5 * time.Second,
		},
		{
			name:      "fallback when protocol has no time between blocks",
			blockTime: 0,
			seconds:   0,
			want:      10 * time.Second,
		},
		{
			name:      "explicit interval without protocol",
			blockTime: 0,
			seconds:   5,
			want:      5 * time.Second,
		},
		{
			name:      "protocol time between blocks after migration",
			blockTime: 15,
			seconds:   0,
			want:      15 * time.Second,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bi := &BlockchainIndexer{Network: types.Mainnet}
			bi.blockTime.Store(tt.blockTime * int64(time.Second))

			require.Equal(t, tt.want, bi.tickerDuration(tt.seconds))
		})
	}
}

func TestSetUpdateTicker(t *testing.T) {
	t.Run("delivers the requested interval to the Start loop", func(t *testing.T) {
		bi := &BlockchainIndexer{
			Network:      types.Mainnet,
			refreshTimer: make(chan time.Duration, 10),
		}
		bi.blockTime.Store(8 * int64(time.Second))

		bi.setUpdateTicker(5)
		bi.setUpdateTicker(0)

		require.Equal(t, 5*time.Second, <-bi.refreshTimer)
		require.Equal(t, 8*time.Second, <-bi.refreshTimer)
	})

	t.Run("drops the update instead of blocking when the queue is full", func(t *testing.T) {
		bi := &BlockchainIndexer{
			Network:      types.Mainnet,
			refreshTimer: make(chan time.Duration, 1),
		}

		done := make(chan struct{})
		go func() {
			defer close(done)
			bi.setUpdateTicker(5)
			bi.setUpdateTicker(5)
		}()

		select {
		case <-done:
		case <-time.After(time.Second):
			t.Fatal("setUpdateTicker blocked on a full refreshTimer")
		}

		require.Len(t, bi.refreshTimer, 1)
	})
}
