package contract

import (
	jsoniter "github.com/json-iterator/go"

	stdJSON "encoding/json"

	"github.com/baking-bad/bcdhub/internal/bcd/ast"
	"github.com/baking-bad/bcdhub/internal/bcd/base"
	"github.com/baking-bad/bcdhub/internal/bcd/consts"
	"github.com/baking-bad/bcdhub/internal/bcd/contract/language"
	"github.com/baking-bad/bcdhub/internal/bcd/types"
)

var json = jsoniter.ConfigCompatibleWithStandardLibrary

type contractData struct {
	Code    stdJSON.RawMessage `json:"code"`
	Storage stdJSON.RawMessage `json:"storage"`
}

// Parser -
type Parser struct {
	Code    *ast.Script
	Storage ast.UntypedAST

	FailStrings        types.Set
	Tags               types.Set
	Annotations        types.Set
	HardcodedAddresses types.Set

	Hash     string
	Language string
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

	hardcoded := findHardcodedAddresses(cd.Code)
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
		Tags:               tags,
		HardcodedAddresses: hardcoded,
		Hash:               hash,
		Language:           language.LangUnknown,
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
	return p.parseStorage()
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
	if p.Code.Code[0].Prim == consts.PrimArray && len(p.Code.Code[0].Args) > 0 {
		lang := language.GetFromFirstPrim(p.Code.Code[0].Args[0])
		p.setLanguage(lang)
	}

	return p.parse(p.Code.Code, p.handleCodeNode)
}

func (p *Parser) parseParameter() error {
	if len(p.Code.Parameter) == 0 {
		return nil
	}

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
	lang := language.DetectMichelsonInStorage(p.Code.Storage[0])
	p.setLanguage(lang)

	return p.parse(p.Code.Storage, p.handleStorageNode)
}

func (p *Parser) handleParameterNode(node *base.Node) error {
	if len(node.Annots) > 0 {
		p.Annotations.Append(filterAnnotations(node.Annots)...)

		lang := language.GetFromParameter(node)
		p.setLanguage(lang)
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
	failString := parseFail(node)
	if failString != "" {
		p.FailStrings.Add(failString)
	}

	if len(node.Annots) > 0 {
		p.Annotations.Append(filterAnnotations(node.Annots)...)
	}

	lang := language.GetFromCode(node)
	p.setLanguage(lang)

	p.Tags.Append(primTags(node))

	return nil
}

func (p *Parser) setLanguage(lang string) {
	if lang == language.LangUnknown || p.Language == lang {
		return
	}
	prevPriority := language.GetPriority(p.Language)
	newPriority := language.GetPriority(lang)
	if newPriority > prevPriority {
		p.Language = lang
	}
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
