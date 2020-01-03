package contractparser

import (
	"fmt"
	"github.com/aopoltorzhicky/bcdhub/internal/tlsh"
	"log"
	"strings"
)

// Code -
type Code struct {
	Parameter []map[string]interface{}
	Storage   []interface{}
	Code      []interface{}

	Hash string

	fuzzyReader *HashReader

	Tags        map[string]struct{}
	Language    string
	FailStrings Set
	Primitives  Set
	Annotations Set
}

// Entrypoint -
type Entrypoint struct {
	Name string
	Args []string
}

func newCode(script map[string]interface{}) (Code, error) {
	res := Code{
		Language:    LangUnknown,
		Tags:        make(map[string]struct{}),
		fuzzyReader: NewHashReader(),
		FailStrings: make(Set, 0),
		Primitives:  make(Set, 0),
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
		if err := res.parseCodeSection(c[i]); err != nil {
			return res, err
		}
	}
	return res, nil
}

func (c *Code) parseCodeSection(section interface{}) error {
	s, ok := section.(map[string]interface{})
	if !ok {
		return fmt.Errorf("Invalid section type")
	}
	prim, ok := s["prim"]
	if !ok {
		return fmt.Errorf("Can't find tag 'prim' in section")
	}
	sPrim := prim.(string)

	args, ok := s["args"]
	if !ok {
		return fmt.Errorf("Can't find tag 'args' in section: %s", sPrim)
	}
	vargs, ok := args.([]interface{})
	if !ok {
		return fmt.Errorf("Invalid type tag 'args' in section: %s", sPrim)
	}

	switch sPrim {
	case "code":
		c.Code = vargs
	case "storage":
		c.Storage = vargs
	case "parameter":
		res := make([]map[string]interface{}, len(vargs))
		for i := range vargs {
			res[i], ok = vargs[i].(map[string]interface{})
			if !ok {
				return fmt.Errorf("Invalid type tag 'args' in section: %s", sPrim)
			}
		}
		c.Parameter = res
	}

	return nil
}

func (c *Code) entrypoints() []Entrypoint {
	e := make([]Entrypoint, 0)
	for _, p := range c.Parameter {
		e = append(e, findEntrypoint(p)...)
	}
	return e
}

func parseEntrypointArgs(args interface{}) []string {
	if args == nil {
		return nil
	}
	return argsToFlat(args)
}

func argsToFlat(args interface{}) []string {
	res := make([]string, 0)
	switch val := args.(type) {
	case []interface{}:
		for i := range val {
			res = append(res, argsToFlat(val[i])...)
		}
	case map[string]interface{}:
		if prim, ok := val["prim"]; ok {
			sPrim := prim.(string)
			if sPrim == "pair" || sPrim == "or" {
				if a, ok := val["args"]; ok {
					res = append(res, argsToFlat(a)...)
				}
			} else {
				if a, ok := val["args"]; ok {
					res = append(res, argsToFlat(a)...)
				} else {
					res = append(res, sPrim)
				}
			}

		}
	}
	return res
}

func findEntrypoint(v map[string]interface{}) []Entrypoint {
	annots, ok := v["annots"]
	if ok {
		ann := annots.([]interface{})
		e := make([]Entrypoint, len(ann))
		for i := range ann {
			name := ann[i].(string)
			name = strings.Trim(name, "%")

			e[i] = Entrypoint{
				Name: name,
				Args: parseEntrypointArgs(v["args"]),
			}
		}
		return e
	}

	vargs, ok := v["args"]
	if !ok {
		return nil
	}
	args := vargs.([]interface{})
	es := make([]Entrypoint, 0)
	for _, a := range args {
		es = append(es, findEntrypoint(a.(map[string]interface{}))...)
	}
	return es
}

func (c *Code) print() {
	log.Print("Entrypoints:")
	entrypoints := c.entrypoints()
	for _, e := range entrypoints {
		log.Println(e.Name)
	}
}

func (c *Code) parseCodePart() error {
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

func (c *Code) handlePrimitive(node *Node) (err error) {
	if node.Prim != "" {
		c.Primitives.Append(node.Prim)
		c.fuzzyReader.WriteString(node.Prim)
	}

	if node.HasAnnots() {
		c.Annotations.Append(node.Annotations...)
	}

	c.detectLanguage(node)
	c.findTags(node)

	return nil
}

func (c *Code) detectLanguage(node *Node) {
	if c.Language != LangUnknown {
		return
	}
	if detectLiquidity(node, c.entrypoints()) {
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

func (c *Code) findTags(node *Node) {
	tag := primTags(node)
	_, ok := c.Tags[tag]
	if tag != "" && !ok {
		c.Tags[tag] = struct{}{}
	}
}
