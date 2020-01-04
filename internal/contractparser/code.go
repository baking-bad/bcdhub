package contractparser

import (
	"fmt"

	"github.com/aopoltorzhicky/bcdhub/internal/tlsh"
)

// Code -
type Code struct {
	Parameter []map[string]interface{}
	Storage   Storage
	Code      []interface{}

	Hash string

	fuzzyReader *HashReader

	Tags        Set
	Language    string
	FailStrings Set
	Primitives  Set
	Annotations Set
	Entrypoints Entrypoints
}

// Entrypoints -
type Entrypoints map[string][]string

func newCode(script map[string]interface{}) (Code, error) {
	res := Code{
		Language:    LangUnknown,
		fuzzyReader: NewHashReader(),
		FailStrings: make(Set),
		Primitives:  make(Set),
		Tags:        make(Set),
		Annotations: make(Set),
		Entrypoints: make(Entrypoints),
	}
	code, ok := script["code"]
	if !ok {
		return res, fmt.Errorf("Can't find tag 'code'")
	}
	c, ok := code.([]interface{})
	if !ok {
		return res, fmt.Errorf("Invalid tag 'code' type")
	}
	if len(c) != 3 {
		return res, fmt.Errorf("Invalid tag 'code' length")
	}

	for i := range c {
		n := newNode(c[i].(map[string]interface{}))
		if err := res.parseStruct(n); err != nil {
			return res, err
		}
	}
	return res, nil
}

func (c *Code) parseStruct(node Node) error {
	switch node.Prim {
	case "CODE":
		c.Code = node.Args
	case "STORAGE":
		store, err := newStorage(node.Args)
		if err != nil {
			return err
		}
		c.Storage = store
	case "PARAMETER":
		res := make([]map[string]interface{}, len(node.Args))
		for i := range node.Args {
			r, ok := node.Args[i].(map[string]interface{})
			if !ok {
				return fmt.Errorf("Unknown parameter type: %T", node.Args[i])
			}
			res = append(res, r)
		}
		c.Parameter = res
	}

	return nil
}

func (c *Code) parse() error {
	if err := c.parseCode(); err != nil {
		return err
	}
	if err := c.parseParameter(); err != nil {
		return err
	}
	return nil
}

func (c *Code) parseCode() error {
	for _, val := range c.Code {
		if err := c.parsePrimitive(val); err != nil {
			return err
		}
	}
	h, err := tlsh.HashReader(c.fuzzyReader)
	if err != nil {
		return err
	}
	c.Hash = h.String()
	return nil
}

func (c *Code) parseParameter() error {
	for i := range c.Parameter {
		c.findEntrypoint(c.Parameter[i])
	}
	return nil
}

func (c *Code) parsePrimitive(val interface{}) error {
	switch t := val.(type) {
	case []interface{}:
		for _, a := range t {
			if err := c.parsePrimitive(a); err != nil {
				return err
			}
		}
		if fail := parseFail(t); fail != nil {
			if fail.With != "" {
				c.FailStrings.Append(fail.With)
			}
		}
	case map[string]interface{}:
		node := newNode(t)
		for i := range node.Args {
			c.parsePrimitive(node.Args[i])
		}
		if err := c.handlePrimitive(node); err != nil {
			return err
		}
	default:
		return fmt.Errorf("Unknown value type: %T", t)
	}
	return nil
}

func (c *Code) handlePrimitive(node Node) (err error) {
	if node.Prim != "" {
		c.Primitives.Append(node.Prim)
		c.fuzzyReader.WriteString(node.Prim)
	}

	if node.HasAnnots() {
		c.Annotations.Append(node.Annotations...)
	}

	c.detectLanguage(node)
	c.Tags.Append(primTags(node))

	return nil
}

func (c *Code) detectLanguage(node Node) {
	if c.Language != LangUnknown {
		return
	}
	if detectLiquidity(node, c.Entrypoints) {
		c.Language = LangLiquidity
		return
	}
	if detectPython(node) {
		c.Language = LangPython
		return
	}
	if detectLIGO(node) {
		c.Language = LangLigo
		return
	}
	if c.Language == "" {
		c.Language = LangUnknown
	}
}
