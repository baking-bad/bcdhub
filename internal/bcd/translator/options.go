package translator

// ConverterOption -
type ConverterOption func(*Converter)

// WithGrammarFile -
func WithGrammarFile(filename string) ConverterOption {
	return func(c *Converter) {
		grammar, err := readFileToString(filename)
		if err != nil {
			c.err = err
			return
		}
		c.grammar = grammar
	}
}

// WithGrammar -
func WithGrammar(grammar string) ConverterOption {
	return func(c *Converter) {
		c.grammar = grammar
	}
}

// WithDefaultGrammar -
func WithDefaultGrammar() ConverterOption {
	return func(c *Converter) {
		c.grammar = defaultGrammar
	}
}

// WithDebug -
func WithDebug() ConverterOption {
	return func(c *Converter) {
		c.debug = true
	}
}
