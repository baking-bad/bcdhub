package transfer

// ParserOption -
type ParserOption func(dp *Parser)

// WithTokenViews -
func WithTokenViews(events TokenEvents) ParserOption {
	return func(tp *Parser) {
		tp.events = events
	}
}
