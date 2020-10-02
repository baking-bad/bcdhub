package tzip

import "time"

// ParserConfig -
type ParserConfig struct {
	IPFSGateways []string
	HTTPTimeout  time.Duration
}
