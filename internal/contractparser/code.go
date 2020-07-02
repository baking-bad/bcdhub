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
	Language    helpers.Set
	FailStrings helpers.Set
	Annotations helpers.Set
	Hash        string
}

func newCode(script gjson.Result) (Code, error) {
	res := Code{
		parser:      &parser{},
		Language:    make(helpers.Set),
		FailStrings: make(helpers.Set),
		Tags:        make(helpers.Set),
		Annotations: make(helpers.Set),
	}
	res.primHandler = res.handlePrimitive
	res.arrayHandler = res.handleArray

	code := script.Get("code").Array()
	if len(code) != 3 {
		return res, fmt.Errorf("Invalid tag 'code' length")
	}

	hash, err := ComputeContractHash(script.Get("code").Raw)
	if err != nil {
		return res, err
	}
	res.Hash = hash

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
	lang := language.GetFromFirstPrim(args.Get("0.0"))
	c.Language.Add(lang)

	for _, val := range args.Array() {
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
	if n.HasAnnots() {
		c.Annotations.Append(filterAnnotations(n.Annotations)...)
	}

	lang := language.GetFromCode(n)
	c.Language.Add(lang)

	c.Tags.Append(primTags(n))

	return nil
}
