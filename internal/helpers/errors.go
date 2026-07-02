package helpers

import (
	"context"
	"errors"
	"net"
	"syscall"
)

// IsTransientError reports whether err is a temporary network failure
// (timeout, refused/reset connection) that is likely to succeed on retry.
func IsTransientError(err error) bool {
	if err == nil {
		return false
	}

	if errors.Is(err, context.DeadlineExceeded) ||
		errors.Is(err, syscall.ECONNREFUSED) ||
		errors.Is(err, syscall.ECONNRESET) {
		return true
	}

	var netErr net.Error
	return errors.As(err, &netErr) && netErr.Timeout()
}
