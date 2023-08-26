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

// WithRateLimit -
func WithRateLimit(requestPerSecond int) NodeOption {
	return func(node *NodeRPC) {
		if requestPerSecond > 0 {
			node.rateLimit = rate.NewLimiter(rate.Every(time.Second/time.Duration(requestPerSecond)), requestPerSecond)
		}
	}
}

// WithLog -
func WithLog() NodeOption {
	return func(node *NodeRPC) {
		node.needLog = true
	}
}
