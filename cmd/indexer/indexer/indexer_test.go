package indexer

import (
	"context"
	"net/url"
	"testing"
	"time"

	"github.com/baking-bad/bcdhub/internal/models/types"
	"github.com/getsentry/sentry-go"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/require"
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
