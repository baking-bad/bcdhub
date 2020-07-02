package noderpc

import "time"

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
