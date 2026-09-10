package periodic

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"github.com/baking-bad/bcdhub/internal/models/types"
	"github.com/baking-bad/bcdhub/internal/teztnets"
	"github.com/stretchr/testify/require"
)

// teztnetsServer serves a teztnets.json whose contents can be swapped between calls.
func teztnetsServer(t *testing.T, info *atomic.Value) *httptest.Server {
	t.Helper()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, "/teztnets.json", r.URL.Path)
		require.NoError(t, json.NewEncoder(w).Encode(info.Load().(teztnets.Info)))
	}))
	t.Cleanup(srv.Close)

	return srv
}

// mustNotBlock runs f and fails the test if it does not return in time. Every call that
// takes the worker mutex goes through it, so a lock left held by checkNetwork surfaces
// as a failure instead of hanging the test binary until the go test timeout.
func mustNotBlock(t *testing.T, name string, f func()) {
	t.Helper()

	done := make(chan struct{})
	go func() {
		defer close(done)
		f()
	}()

	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatalf("%s blocked: the worker mutex was left locked", name)
	}
}

func TestWorkerCheckNetwork(t *testing.T) {
	var info atomic.Value
	info.Store(teztnets.Info{
		"mainnet":  {RPCURL: "https://rpc.example.com/one"},
		"ghostnet": {RPCURL: "https://rpc.example.com/other-network"},
	})
	srv := teztnetsServer(t, &info)

	var handled []string
	w, err := New(
		Config{InfoBaseURL: srv.URL},
		types.Mainnet,
		func(ctx context.Context, network, newUrl string) error {
			handled = append(handled, newUrl)
			return nil
		},
	)
	require.NoError(t, err)

	ctx := context.Background()

	check := func(t *testing.T) bool {
		t.Helper()

		var (
			changed bool
			err     error
		)
		mustNotBlock(t, "checkNetwork", func() { changed, err = w.checkNetwork(ctx) })
		require.NoError(t, err)
		return changed
	}

	url := func(t *testing.T) string {
		t.Helper()

		var url string
		mustNotBlock(t, "URL()", func() { url = w.URL() })
		return url
	}

	t.Run("first observation is adopted without calling the handler", func(t *testing.T) {
		require.True(t, check(t))
		require.Empty(t, handled)
		require.Equal(t, "https://rpc.example.com/one", url(t))
	})

	t.Run("unchanged url is a no-op", func(t *testing.T) {
		require.False(t, check(t))
		require.Empty(t, handled)
		require.Equal(t, "https://rpc.example.com/one", url(t))
	})

	t.Run("new url calls the handler and is adopted", func(t *testing.T) {
		info.Store(teztnets.Info{"mainnet": {RPCURL: "https://rpc.example.com/two"}})

		require.True(t, check(t))
		require.Equal(t, []string{"https://rpc.example.com/two"}, handled)
		require.Equal(t, "https://rpc.example.com/two", url(t))
	})
}

// TestWorkerURLDuringHandler pins that the worker does not hold its own mutex while the
// change handler runs: handleUrlChanged waits for the indexer loop to exit and then
// re-initialises it, and URL() must stay answerable for that whole time.
func TestWorkerURLDuringHandler(t *testing.T) {
	var info atomic.Value
	info.Store(teztnets.Info{"mainnet": {RPCURL: "https://rpc.example.com/one"}})
	srv := teztnetsServer(t, &info)

	var w *Worker
	seen := make(chan string, 1)
	w, err := New(
		Config{InfoBaseURL: srv.URL},
		types.Mainnet,
		func(ctx context.Context, network, newUrl string) error {
			var url string
			mustNotBlock(t, "URL() inside the handler", func() { url = w.URL() })
			seen <- url
			return nil
		},
	)
	require.NoError(t, err)

	ctx := context.Background()
	mustNotBlock(t, "checkNetwork", func() { _, err = w.checkNetwork(ctx) })
	require.NoError(t, err)

	info.Store(teztnets.Info{"mainnet": {RPCURL: "https://rpc.example.com/two"}})
	mustNotBlock(t, "checkNetwork", func() { _, err = w.checkNetwork(ctx) })
	require.NoError(t, err)

	require.Equal(t, "https://rpc.example.com/one", <-seen)
}

func TestGeneralWorkerCheckNetwork(t *testing.T) {
	var info atomic.Value
	info.Store(teztnets.Info{
		"mainnet":  {RPCURL: "https://rpc.example.com/mainnet"},
		"ghostnet": {RPCURL: "https://rpc.example.com/ghostnet"},
	})
	srv := teztnetsServer(t, &info)

	handled := make(map[string]string)
	w, err := NewGeneralWorker(
		Config{InfoBaseURL: srv.URL},
		func(ctx context.Context, network, newUrl string) error {
			handled[network] = newUrl
			return nil
		},
	)
	require.NoError(t, err)

	ctx := context.Background()

	check := func(t *testing.T) bool {
		t.Helper()

		var (
			changed bool
			err     error
		)
		mustNotBlock(t, "checkNetwork", func() { changed, err = w.checkNetwork(ctx) })
		require.NoError(t, err)
		return changed
	}

	urls := func(t *testing.T) map[string]string {
		t.Helper()

		var urls map[string]string
		mustNotBlock(t, "URLs()", func() { urls = w.URLs() })
		return urls
	}

	// Unlike Worker, GeneralWorker has no "skip the handler on the first observation"
	// guard, so every network it learns about is reported to the handler right away.
	t.Run("every network is adopted and handled on the first observation", func(t *testing.T) {
		require.True(t, check(t))
		require.Equal(t, map[string]string{
			"mainnet":  "https://rpc.example.com/mainnet",
			"ghostnet": "https://rpc.example.com/ghostnet",
		}, handled)
		require.Equal(t, map[string]string{
			"mainnet":  "https://rpc.example.com/mainnet",
			"ghostnet": "https://rpc.example.com/ghostnet",
		}, urls(t))
	})

	t.Run("unchanged urls are a no-op", func(t *testing.T) {
		require.False(t, check(t))
		require.Equal(t, map[string]string{
			"mainnet":  "https://rpc.example.com/mainnet",
			"ghostnet": "https://rpc.example.com/ghostnet",
		}, handled)
	})

	t.Run("only the changed network is handled", func(t *testing.T) {
		info.Store(teztnets.Info{
			"mainnet":  {RPCURL: "https://rpc.example.com/mainnet"},
			"ghostnet": {RPCURL: "https://rpc.example.com/ghostnet-2"},
		})

		require.True(t, check(t))
		require.Equal(t, map[string]string{
			"mainnet":  "https://rpc.example.com/mainnet",
			"ghostnet": "https://rpc.example.com/ghostnet-2",
		}, handled)
	})

	t.Run("URLs returns a copy", func(t *testing.T) {
		got := urls(t)
		got["mainnet"] = "mutated"
		delete(got, "ghostnet")

		require.Equal(t, map[string]string{
			"mainnet":  "https://rpc.example.com/mainnet",
			"ghostnet": "https://rpc.example.com/ghostnet-2",
		}, urls(t))
	})
}

// TestWorkerRetriesAfterHandlerError pins that a failed switch is retried: the worker
// must not adopt a url it could not apply, otherwise the network silently stops being
// indexed behind a single log line.
func TestWorkerRetriesAfterHandlerError(t *testing.T) {
	var info atomic.Value
	info.Store(teztnets.Info{"mainnet": {RPCURL: "https://rpc.example.com/one"}})
	srv := teztnetsServer(t, &info)

	var fail atomic.Bool
	fail.Store(true)

	var calls int
	w, err := New(
		Config{InfoBaseURL: srv.URL},
		types.Mainnet,
		func(ctx context.Context, network, newUrl string) error {
			calls++
			if fail.Load() {
				return errors.New("reinit failed")
			}
			return nil
		},
	)
	require.NoError(t, err)

	ctx := context.Background()

	check := func(t *testing.T) bool {
		t.Helper()

		var (
			changed bool
			err     error
		)
		mustNotBlock(t, "checkNetwork", func() { changed, err = w.checkNetwork(ctx) })
		require.NoError(t, err)
		return changed
	}

	url := func(t *testing.T) string {
		t.Helper()

		var url string
		mustNotBlock(t, "URL()", func() { url = w.URL() })
		return url
	}

	// first observation is adopted without the handler
	require.True(t, check(t))
	require.Zero(t, calls)
	require.Equal(t, "https://rpc.example.com/one", url(t))

	info.Store(teztnets.Info{"mainnet": {RPCURL: "https://rpc.example.com/two"}})

	// the handler fails, so the old url stays in place and nothing is reported as changed
	require.False(t, check(t))
	require.Equal(t, 1, calls)
	require.Equal(t, "https://rpc.example.com/one", url(t))

	// the next poll tries again
	fail.Store(false)
	require.True(t, check(t))
	require.Equal(t, 2, calls)
	require.Equal(t, "https://rpc.example.com/two", url(t))
}

// TestGeneralWorkerRetriesAfterHandlerError is the GeneralWorker counterpart: a network
// whose handler failed keeps its old url and is retried, while the others are unaffected.
func TestGeneralWorkerRetriesAfterHandlerError(t *testing.T) {
	var info atomic.Value
	info.Store(teztnets.Info{
		"mainnet":  {RPCURL: "https://rpc.example.com/mainnet"},
		"ghostnet": {RPCURL: "https://rpc.example.com/ghostnet"},
	})
	srv := teztnetsServer(t, &info)

	var failing atomic.Value
	failing.Store("")

	w, err := NewGeneralWorker(
		Config{InfoBaseURL: srv.URL},
		func(ctx context.Context, network, newUrl string) error {
			if failing.Load().(string) == network {
				return errors.New("reinit failed")
			}
			return nil
		},
	)
	require.NoError(t, err)

	ctx := context.Background()

	check := func(t *testing.T) bool {
		t.Helper()

		var (
			changed bool
			err     error
		)
		mustNotBlock(t, "checkNetwork", func() { changed, err = w.checkNetwork(ctx) })
		require.NoError(t, err)
		return changed
	}

	urls := func(t *testing.T) map[string]string {
		t.Helper()

		var urls map[string]string
		mustNotBlock(t, "URLs()", func() { urls = w.URLs() })
		return urls
	}

	require.True(t, check(t))

	failing.Store("ghostnet")
	info.Store(teztnets.Info{
		"mainnet":  {RPCURL: "https://rpc.example.com/mainnet-2"},
		"ghostnet": {RPCURL: "https://rpc.example.com/ghostnet-2"},
	})

	require.True(t, check(t))
	require.Equal(t, map[string]string{
		"mainnet":  "https://rpc.example.com/mainnet-2",
		"ghostnet": "https://rpc.example.com/ghostnet",
	}, urls(t))

	failing.Store("")
	require.True(t, check(t))
	require.Equal(t, map[string]string{
		"mainnet":  "https://rpc.example.com/mainnet-2",
		"ghostnet": "https://rpc.example.com/ghostnet-2",
	}, urls(t))
}
