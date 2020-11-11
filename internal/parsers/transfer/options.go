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
