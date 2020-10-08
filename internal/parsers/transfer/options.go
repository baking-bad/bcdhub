package transfer

// ParserOption -
type ParserOption func(dp *Parser)

// WithTokenViews -
func WithTokenViews(views TokenViews) ParserOption {
	return func(tp *Parser) {
		tp.views = views
	}
}
