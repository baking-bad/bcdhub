package transfer

import "github.com/baking-bad/bcdhub/internal/parsers/stacktrace"

// ParserOption -
type ParserOption func(dp *Parser)

// WithStackTrace -
func WithStackTrace(stackTrace *stacktrace.StackTrace) ParserOption {
	return func(dp *Parser) {
		dp.stackTrace = stackTrace
	}
}

// WithNetwork -
func WithNetwork(network string) ParserOption {
	return func(dp *Parser) {
		dp.network = network
	}
}

// WithChainID -
func WithChainID(chainID string) ParserOption {
	return func(dp *Parser) {
		dp.chainID = chainID
	}
}

// WithGasLimit -
func WithGasLimit(gasLimit int64) ParserOption {
	return func(dp *Parser) {
		dp.gasLimit = gasLimit
	}
}
