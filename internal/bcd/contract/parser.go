package contract

import (
	jsoniter "github.com/json-iterator/go"

	stdJSON "encoding/json"

	"github.com/baking-bad/bcdhub/internal/bcd/ast"
	"github.com/baking-bad/bcdhub/internal/bcd/base"
	"github.com/baking-bad/bcdhub/internal/bcd/consts"
	"github.com/baking-bad/bcdhub/internal/bcd/types"
)

var json = jsoniter.ConfigCompatibleWithStandardLibrary

type contractData struct {
	Code    stdJSON.RawMessage `json:"code"`
	Storage stdJSON.RawMessage `json:"storage"`
}

// Fingerprint -
type Fingerprint struct {
	Parameter []byte
	Storage   []byte
	Code      []byte
}

// Parser -
type Parser struct {
	Code    *ast.Script
	Storage ast.UntypedAST

	FailStrings        types.Set
	Tags               types.Set
	Annotations        types.Set
	HardcodedAddresses types.Set
	Constants          types.Set

	Fingerprint Fingerprint

	Hash string

	CodeRaw []byte
}

// NewParser -
func NewParser(data []byte) (*Parser, error) {
	var cd contractData
	if err := json.Unmarshal(data, &cd); err != nil {
		return nil, err
	}

	script, err := ast.NewScript(cd.Code)
	if err != nil {
		return nil, err
	}

	var storage ast.UntypedAST
	if err := json.Unmarshal(cd.Storage, &storage); err != nil {
		return nil, err
	}

	hash, err := ComputeHash(cd.Code)
	if err != nil {
		return nil, err
	}

	tags := make(types.Set)

	if isDelegatorContract(cd.Code, storage) {
		tags.Add(consts.DelegatorTag)
	}

	if isMultisigContract(cd.Code) {
		tags.Add(consts.MultisigTag)
	}

	return &Parser{
		Code:               script,
		Storage:            storage,
		FailStrings:        make(types.Set),
		Annotations:        make(types.Set),
		Constants:          make(types.Set),
		HardcodedAddresses: make(types.Set),
		Tags:               tags,
		Hash:               hash,
		Fingerprint:        Fingerprint{},
		CodeRaw:            cd.Code,
	}, nil
}

// Parse -
func (p *Parser) Parse() error {
	if err := p.parseCode(); err != nil {
		return err
	}
	if err := p.parseParameter(); err != nil {
		return err
	}
	if err := p.parseStorage(); err != nil {
		return err
	}

	return p.setStorageTypeTags()
}

// FindConstants -
func (p *Parser) FindConstants() ([]string, error) {
	if len(p.Code.Code) == 0 {
		return nil, nil
	}

	if err := p.parse(p.Code.Code, p.handleConstants); err != nil {
		return nil, err
	}

	if err := p.parse(p.Code.Storage, p.handleConstants); err != nil {
		return nil, err
	}

	if err := p.parse(p.Code.Parameter, p.handleConstants); err != nil {
		return nil, err
	}

	return p.Constants.Values(), nil
}

// IsUpgradable -
func (p *Parser) IsUpgradable() bool {
	for _, params := range p.Code.Parameter {
		if params.Prim != consts.LAMBDA {
			continue
		}

		for _, s := range p.Code.Storage {
			if s.Prim != consts.LAMBDA {
				continue
			}

			if params.Compare(s) {
				return true
			}
		}
	}
	return false
}

func (p *Parser) setStorageTypeTags() error {
	storageTyp, err := p.Code.StorageType()
	if err != nil {
		return err
	}

	if p.Annotations.Has("%ledger") {
		if isNftLedger(storageTyp) {
			p.Tags.Add(consts.LedgerTag)
		}
	}

	return nil
}

func (p *Parser) parse(tree ast.UntypedAST, handler func(node *base.Node) error) error {
	for i := range tree {
		if err := handler(tree[i]); err != nil {
			return err
		}
		if len(tree[i].Args) > 0 {
			if err := p.parse(tree[i].Args, handler); err != nil {
				return err
			}
		}
	}
	return nil
}

func (p *Parser) parseCode() error {
	if len(p.Code.Code) == 0 {
		return nil
	}
	f, err := p.Code.Code.Fingerprint(true)
	if err != nil {
		return err
	}
	p.Fingerprint.Code = f

	return p.parse(p.Code.Code, p.handleCodeNode)
}

func (p *Parser) parseParameter() error {
	if len(p.Code.Parameter) == 0 {
		return nil
	}

	f, err := p.Code.Parameter.Fingerprint(false)
	if err != nil {
		return err
	}
	p.Fingerprint.Parameter = f

	typedParamTree, err := p.Code.Parameter.ToTypedAST()
	if err != nil {
		return err
	}
	tags := ast.FindContractInterfaces(typedParamTree)
	p.Tags.Append(tags...)

	return p.parse(p.Code.Parameter, p.handleParameterNode)
}

func (p *Parser) parseStorage() error {
	if len(p.Code.Storage) == 0 {
		return nil
	}

	f, err := p.Code.Storage.Fingerprint(false)
	if err != nil {
		return err
	}
	p.Fingerprint.Storage = f
	return p.parse(p.Code.Storage, p.handleStorageNode)
}

func (p *Parser) handleParameterNode(node *base.Node) error {
	if len(node.Annots) > 0 {
		p.Annotations.Append(filterAnnotations(node.Annots)...)
	}

	return nil
}

func (p *Parser) handleStorageNode(node *base.Node) error {
	if len(node.Annots) > 0 {
		p.Annotations.Append(filterAnnotations(node.Annots)...)
	}

	return nil
}

func (p *Parser) handleCodeNode(node *base.Node) error {
	if constant := parseConstants(node); constant != "" {
		p.Constants.Add(constant)
	} else if failString := parseFail(node); failString != "" {
		p.FailStrings.Add(failString)
	}

	if hardcoded := parseHardcoded(node); len(hardcoded) > 0 {
		p.HardcodedAddresses.Append(hardcoded...)
	}

	if len(node.Annots) > 0 {
		p.Annotations.Append(filterAnnotations(node.Annots)...)
	}

	p.Tags.Append(primTags(node))

	return nil
}

func (p *Parser) handleConstants(node *base.Node) error {
	if constant := parseConstants(node); constant != "" {
		p.Constants.Add(constant)
	}
	return nil
}

func parseHardcoded(node *base.Node) []string {
	if node.Prim != consts.STRING || node.StringValue == nil {
		return nil
	}
	value := *node.StringValue
	if len(value) < 36 {
		return nil
	}
	return regAddress.FindAllString(value, -1)
}

func parseConstants(node *base.Node) string {
	if node == nil {
		return ""
	}

	if node.Prim != consts.CONSTANT || len(node.Args) != 1 {
		return ""
	}

	arg := node.Args[0]
	if arg.StringValue == nil {
		return ""
	}

	value := arg.StringValue
	return *value
}

func parseFail(node *base.Node) string {
	if node.Prim != consts.PrimArray || len(node.Args) < 2 {
		return ""
	}

	var pushArgs []*base.Node
	var hasFailWith bool
	for i := range node.Args {
		switch {
		case !hasFailWith && node.Args[i].Prim == "FAILWITH":
			hasFailWith = true
		case len(pushArgs) == 0 && node.Args[i].Prim == "PUSH":
			pushArgs = node.Args[i].Args
		}
	}

	if hasFailWith {
		for i := range pushArgs {
			if pushArgs[i].StringValue != nil {
				return *pushArgs[i].StringValue
			}
		}
	}

	return ""
}
