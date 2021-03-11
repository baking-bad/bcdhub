package contract

import (
	"github.com/baking-bad/bcdhub/internal/bcd"
	"github.com/baking-bad/bcdhub/internal/bcd/consts"
	astContract "github.com/baking-bad/bcdhub/internal/bcd/contract"
	"github.com/baking-bad/bcdhub/internal/helpers"
	"github.com/baking-bad/bcdhub/internal/models"
	"github.com/baking-bad/bcdhub/internal/models/contract"
	"github.com/baking-bad/bcdhub/internal/models/operation"
	"github.com/pkg/errors"
)

// Parser -
type Parser struct {
	scriptSaver ScriptSaver
}

// NewParser -
func NewParser(opts ...ParserOption) *Parser {
	parser := &Parser{}
	for i := range opts {
		opts[i](parser)
	}

	return parser
}

// ParserOption -
type ParserOption func(p *Parser)

// WithShareDir -
func WithShareDir(dir string) ParserOption {
	return func(p *Parser) {
		if dir == "" {
			return
		}
		p.scriptSaver = NewFileScriptSaver(dir)
	}
}

// Parse -
func (p *Parser) Parse(operation *operation.Operation) ([]models.Model, error) {
	if !helpers.StringInArray(operation.Kind, []string{
		consts.Origination, consts.OriginationNew, consts.Migration,
	}) {
		return nil, errors.Errorf("Invalid operation kind in computeContractMetrics: %s", operation.Kind)
	}
	contract := contract.Contract{
		Network:    operation.Network,
		Level:      operation.Level,
		Timestamp:  operation.Timestamp,
		Manager:    operation.Source,
		Address:    operation.Destination,
		Delegate:   operation.Delegate,
		LastAction: operation.Timestamp,
	}

	if err := p.computeMetrics(operation, &contract); err != nil {
		return nil, err
	}

	return []models.Model{&contract}, nil
}

func (p *Parser) computeMetrics(operation *operation.Operation, c *contract.Contract) error {
	script, err := astContract.NewParser(operation.Script)
	if err != nil {
		return errors.Errorf("ast.NewScript: %v", err)
	}
	if err := script.Parse(); err != nil {
		return err
	}
	operation.Script = script.CodeRaw

	c.Language = script.Language
	c.Hash = script.Hash
	c.FailStrings = script.FailStrings.Values()
	c.Annotations = script.Annotations.Values()
	c.Tags = script.Tags.Values()
	c.Hardcoded = script.HardcodedAddresses.Values()
	c.Fingerprint = new(contract.Fingerprint)
	c.Fingerprint.Code = script.Fingerprint.Code
	c.Fingerprint.Parameter = script.Fingerprint.Parameter
	c.Fingerprint.Storage = script.Fingerprint.Storage

	params, err := script.Code.Parameter.ToTypedAST()
	if err != nil {
		return err
	}
	c.Entrypoints = params.GetEntrypoints()

	if script.IsUpgradable() {
		c.Tags = append(c.Tags, consts.UpgradableTag)
	}

	protoSymLink, err := bcd.GetProtoSymLink(operation.Protocol)
	if err != nil {
		return err
	}

	if p.scriptSaver != nil {
		return p.scriptSaver.Save(operation.Script, ScriptSaveContext{
			Network: c.Network,
			Address: c.Address,
			Hash:    c.Hash,
			SymLink: protoSymLink,
		})
	}
	return nil
}
