package translator

import (
	"os"
	"regexp"

	"github.com/rs/zerolog/log"
	"github.com/yhirose/go-peg"
)

// Converter -
type Converter struct {
	parser  *peg.Parser
	grammar string
	err     error
	debug   bool
}

// NewConverter -
func NewConverter(opts ...ConverterOption) (Converter, error) {
	c := Converter{}
	for i := range opts {
		opts[i](&c)
	}
	if c.err != nil {
		return c, c.err
	}

	if c.grammar == "" {
		c.grammar = defaultGrammar
	}

	parser, err := peg.NewParser(c.grammar)
	if err != nil {
		return c, err
	}

	if err := parser.EnableAst(); err != nil {
		return c, err
	}
	c.parser = parser
	return c, nil
}

// FromFile -
func (c Converter) FromFile(filename string) (string, error) {
	c.trace()

	michelson, err := readFileToString(filename)
	if err != nil {
		return "", err
	}

	return c.FromString(michelson)
}

// FromString -
func (c Converter) FromString(input string) (string, error) {
	c.trace()

	input = removeComments(input)

	ast, err := c.parser.ParseAndGetAst(input, nil)
	if err != nil {
		return "", err
	}

	return NewJSONTranslator().Translate(ast)
}

func (c Converter) trace() {
	if c.debug {
		c.parser.TracerEnter = func(name string, s string, v *peg.Values, d peg.Any, p int) {
			log.Info().Msgf("Enter: %s %d %d %s", name, p, len(s), s[p:])
		}
		c.parser.TracerLeave = func(name string, s string, v *peg.Values, d peg.Any, p int, l int) {
			if l != -1 {
				log.Info().Msgf("Leave: %s %d %d", name, len(s), l+p)
			}
		}
	}
}

func readFileToString(filename string) (string, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

var oneLineComment = regexp.MustCompile("#[^\n]*")
var multiLineComment = regexp.MustCompile(`\/\*[^\*]*\*/`)

func removeComments(data string) string {
	data = oneLineComment.ReplaceAllString(data, "")
	return multiLineComment.ReplaceAllString(data, "")
}
