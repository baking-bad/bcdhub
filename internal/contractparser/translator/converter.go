package translator

import (
	"io/ioutil"
	"os"

	"github.com/baking-bad/bcdhub/internal/logger"
	"github.com/valyala/fastjson"
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
func (c Converter) FromFile(filename string) (*fastjson.Value, error) {
	c.trace()

	michelson, err := readFileToString(filename)
	if err != nil {
		return nil, err
	}

	return c.FromString(michelson)
}

// FromString -
func (c Converter) FromString(input string) (*fastjson.Value, error) {
	c.trace()

	ast, err := c.parser.ParseAndGetAst(input, nil)
	if err != nil {
		return nil, err
	}

	return NewJSONTranslator().Translate(ast)
}

func (c Converter) trace() {
	if c.debug {
		c.parser.TracerEnter = func(name string, s string, v *peg.Values, d peg.Any, p int) {
			logger.Info("Enter: %s %d %d", name, p, len(s))
		}
		c.parser.TracerLeave = func(name string, s string, v *peg.Values, d peg.Any, p int, l int) {
			if l != -1 {
				logger.Info("Leave: %s %d %d", name, len(s), l+p)
			}
		}
	}
}

func readFileToString(filename string) (string, error) {
	f, err := os.Open(filename)
	if err != nil {
		return "", err
	}
	data, err := ioutil.ReadAll(f)
	if err != nil {
		return "", err
	}
	return string(data), nil
}
