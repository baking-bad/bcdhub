package helpers

import (
	"testing"

	"github.com/getsentry/sentry-go"
	"github.com/stretchr/testify/require"
)

func TestLocalCatchPanicSentry(t *testing.T) {
	t.Run("nil hub does not propagate panic", func(t *testing.T) {
		require.NotPanics(t, func() {
			func() {
				defer LocalCatchPanicSentry(nil)
				panic("boom")
			}()
		})
	})

	t.Run("hub without client does not propagate panic", func(t *testing.T) {
		hub := sentry.NewHub(nil, sentry.NewScope())

		require.NotPanics(t, func() {
			func() {
				defer LocalCatchPanicSentry(hub)
				panic("boom")
			}()
		})
	})

	t.Run("no panic is a no-op", func(t *testing.T) {
		called := false
		func() {
			defer LocalCatchPanicSentry(nil)
			called = true
		}()
		require.True(t, called)
	})
}
