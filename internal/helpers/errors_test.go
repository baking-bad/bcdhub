package helpers

import (
	"context"
	"net/url"
	"syscall"
	"testing"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/require"
)

type timeoutError struct{}

func (timeoutError) Error() string   { return "i/o timeout" }
func (timeoutError) Timeout() bool   { return true }
func (timeoutError) Temporary() bool { return true }

func TestIsTransientError(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{
			name: "nil",
			err:  nil,
			want: false,
		},
		{
			name: "context deadline exceeded",
			err:  context.DeadlineExceeded,
			want: true,
		},
		{
			name: "wrapped url.Error with deadline",
			err: &url.Error{
				Op:  "Get",
				URL: "https://rpc.tzkt.io/archive/chains/main/blocks/head/header",
				Err: context.DeadlineExceeded,
			},
			want: true,
		},
		{
			name: "net.Error timeout",
			err:  timeoutError{},
			want: true,
		},
		{
			name: "connection refused",
			err:  errors.Wrap(syscall.ECONNREFUSED, "dial tcp"),
			want: true,
		},
		{
			name: "connection reset",
			err:  errors.Wrap(syscall.ECONNRESET, "read tcp"),
			want: true,
		},
		{
			name: "regular error",
			err:  errors.New("not found"),
			want: false,
		},
		{
			name: "context canceled",
			err:  context.Canceled,
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.Equal(t, tt.want, IsTransientError(tt.err))
		})
	}
}
