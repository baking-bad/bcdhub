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
