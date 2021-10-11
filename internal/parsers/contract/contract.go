package contract

import (
	"bytes"
	"fmt"

	astContract "github.com/baking-bad/bcdhub/internal/bcd/contract"
	"github.com/baking-bad/bcdhub/internal/config"
	"github.com/baking-bad/bcdhub/internal/models/contract"
	"github.com/baking-bad/bcdhub/internal/models/operation"
	"github.com/baking-bad/bcdhub/internal/models/types"
	"github.com/baking-bad/bcdhub/internal/parsers"
	"github.com/pkg/errors"
)

// Parser -
type Parser struct {
	scriptSaver ScriptSaver
	ctx         *config.Context
}

// NewParser -
func NewParser(ctx *config.Context, opts ...ParserOption) *Parser {
	parser := &Parser{ctx: ctx}
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
func (p *Parser) Parse(operation *operation.Operation) (*parsers.Result, error) {
	if !operation.IsOrigination() {
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

	result := parsers.NewResult()
	result.Contracts = append(result.Contracts, &contract)

	return result, nil
}

func (p *Parser) computeMetrics(operation *operation.Operation, c *contract.Contract) error {
	script, err := astContract.NewParser(operation.Script)
	if err != nil {
		return errors.Wrap(err, "astContract.NewParser")
	}

	constants, err := script.FindConstants()
	if err != nil {
		return errors.Wrap(err, "script.FindConstants")
	}

	if len(constants) > 0 {
		globalConstants, err := p.ctx.GlobalConstants.All(c.Network, constants...)
		if err != nil {
			return err
		}
		c.Constants = globalConstants
		p.replaceConstants(c, operation)

		script, err = astContract.NewParser(operation.Script)
		if err != nil {
			return errors.Wrap(err, "astContract.NewParser")
		}

	}

	if err := script.Parse(); err != nil {
		return err
	}
	operation.Script = script.CodeRaw
	operation.AST = script.Code

	c.Language = script.Language
	c.Hash = script.Hash
	c.FailStrings = script.FailStrings.Values()
	c.Annotations = script.Annotations.Values()
	c.Tags = types.NewTags(script.Tags.Values())
	c.Hardcoded = script.HardcodedAddresses.Values()
	c.FingerprintCode = script.Fingerprint.Code
	c.FingerprintParameter = script.Fingerprint.Parameter
	c.FingerprintStorage = script.Fingerprint.Storage

	params, err := script.Code.Parameter.ToTypedAST()
	if err != nil {
		return err
	}
	c.Entrypoints = params.GetEntrypoints()

	if script.IsUpgradable() {
		c.Tags.Set(types.UpgradableTag)
	}

	c.ProjectID, err = p.ctx.CachedProjectIDByHash(c.Hash)
	if err != nil {
		return err
	}

	proto, err := p.ctx.CachedProtocolByID(operation.Network, operation.ProtocolID)
	if err != nil {
		return err
	}

	if p.scriptSaver != nil {
		return p.scriptSaver.Save(operation.Script, ScriptSaveContext{
			Network: c.Network.String(),
			Address: c.Address,
			Hash:    c.Hash,
			SymLink: proto.SymLink,
		})
	}
	return nil
}

func (p *Parser) replaceConstants(c *contract.Contract, operation *operation.Operation) {
	pattern := `{"prim":"constant","args":[{"string":"%s"}]}`
	for i := range c.Constants {
		operation.Script = bytes.ReplaceAll(
			operation.Script,
			[]byte(fmt.Sprintf(pattern, c.Constants[i].Address)),
			c.Constants[i].Value,
		)
	}
}
