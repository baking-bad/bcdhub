package noderpc

import (
	"time"

	"golang.org/x/time/rate"
)

// NodeOption -
type NodeOption func(*NodeRPC)

// WithTimeout -
func WithTimeout(timeout time.Duration) NodeOption {
	return func(node *NodeRPC) {
		node.timeout = timeout
	}
}

// WithRetryCount -
func WithRetryCount(retryCount int) NodeOption {
	return func(node *NodeRPC) {
		node.retryCount = retryCount
	}
}

// WithRateLimit -
func WithRateLimit(requestPerSecond int) NodeOption {
	return func(node *NodeRPC) {
		if requestPerSecond > 0 {
			node.rateLimit = rate.NewLimiter(rate.Every(time.Second/time.Duration(requestPerSecond)), requestPerSecond)
		}
	}
}
