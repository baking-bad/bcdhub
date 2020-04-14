package contractparser

import (
	"fmt"

	"github.com/baking-bad/bcdhub/internal/contractparser/consts"
	"github.com/baking-bad/bcdhub/internal/contractparser/language"
	"github.com/baking-bad/bcdhub/internal/contractparser/node"
	"github.com/baking-bad/bcdhub/internal/helpers"
	"github.com/tidwall/gjson"
)

// Code -
type Code struct {
	*parser

	Parameter Parameter
	Storage   Storage
	Code      gjson.Result

	Tags        helpers.Set
	Language    string
	FailStrings helpers.Set
	Primitives  helpers.Set
	Annotations helpers.Set
}

func newCode(script gjson.Result) (Code, error) {
	res := Code{
		parser:      &parser{},
		Language:    language.LangUnknown,
		FailStrings: make(helpers.Set),
		Primitives:  make(helpers.Set),
		Tags:        make(helpers.Set),
		Annotations: make(helpers.Set),
	}
	res.primHandler = res.handlePrimitive
	res.arrayHandler = res.handleArray

	code := script.Get("code").Array()
	if len(code) != 3 {
		return res, fmt.Errorf("Invalid tag 'code' length")
	}

	for i := range code {
		n := node.NewNodeJSON(code[i])
		if err := res.parseStruct(n); err != nil {
			return res, err
		}
	}
	return res, nil
}

func (c *Code) parseStruct(n node.Node) error {
	switch n.Prim {
	case consts.CODE:
		c.Code = n.Args
		if err := c.parseCode(n.Args); err != nil {
			return err
		}
	case consts.STORAGE:
		store, err := newStorage(n.Args)
		if err != nil {
			return err
		}
		c.Storage = store
	case consts.PARAMETER:
		parameter, err := newParameter(n.Args)
		if err != nil {
			return err
		}
		c.Parameter = parameter
	}

	return nil
}

func (c *Code) parseCode(args gjson.Result) error {
	for i, val := range args.Array() {
		if i == 0 {
			c.Language = detectLorentzByCast(val)
		}

		if err := c.parse(val); err != nil {
			return err
		}
	}

	return nil
}

func (c *Code) handleArray(arr gjson.Result) error {
	if fail := parseFail(arr); fail != nil {
		if fail.With != "" {
			c.FailStrings.Append(fail.With)
		}
	}
	return nil
}

func (c *Code) handlePrimitive(n node.Node) (err error) {
	c.Primitives.Append(n.Prim)

	if n.HasAnnots() {
		c.Annotations.Append(filterAnnotations(n.Annotations)...)
	}

	if n.Is("") && n.Type == consts.KeyString && c.Language == language.LangUnknown {
		c.Language = language.Get(n)
	}
	c.Tags.Append(primTags(n))

	return nil
}

func detectLorentzByCast(val gjson.Result) string {
	// args.Array()[0][0]["prim"] == "CAST"
	if val.IsArray() {
		if val.Array()[0].IsObject() {
			node := node.NewNodeJSON(val)
			if node.Prim == "CAST" {
				return language.LangLorentz
			}
		}
	}

	return language.LangUnknown
}
