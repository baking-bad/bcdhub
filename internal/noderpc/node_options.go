package noderpc

import (
	"path"
	"time"
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

// WithCache -
func WithCache(cacheDir, network string) NodeOption {
	return func(node *NodeRPC) {
		node.cacheDir = path.Join(cacheDir, "rpc", network)
	}
}
