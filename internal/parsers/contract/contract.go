package contract

import (
	"bytes"
	"fmt"

	"github.com/baking-bad/bcdhub/internal/bcd"
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
	ctx *config.Context
}

// NewParser -
func NewParser(ctx *config.Context) *Parser {
	return &Parser{ctx: ctx}
}

// Parse -
func (p *Parser) Parse(operation *operation.Operation) (*parsers.Result, error) {
	if !operation.IsOrigination() {
		return nil, errors.Errorf("Invalid operation kind in computeContractMetrics: %s", operation.Kind)
	}

	contract := contract.Contract{
		Network:   operation.Network,
		Level:     operation.Level,
		Timestamp: operation.Timestamp,
		Manager:   types.NewNullString(&operation.Source),
		Address:   operation.Destination,
		Delegate: types.NullString{
			Str:   operation.Delegate,
			Valid: operation.Delegate != "",
		},
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

	contractScript, err := p.ctx.Scripts.ByHash(script.Hash)
	if err != nil {
		if !p.ctx.Storage.IsRecordNotFound(err) {
			return err
		}
		contractScript = contract.Script{
			Hash: script.Hash,
			Code: script.CodeRaw,
		}
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
		contractScript.Constants = globalConstants
		p.replaceConstants(&contractScript, operation)

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

	contractScript.FingerprintParameter = script.Fingerprint.Parameter
	contractScript.FingerprintCode = script.Fingerprint.Code
	contractScript.FingerprintStorage = script.Fingerprint.Storage
	contractScript.FailStrings = script.FailStrings.Values()
	contractScript.Annotations = script.Annotations.Values()
	contractScript.Tags = types.NewTags(script.Tags.Values())
	contractScript.Hardcoded = script.HardcodedAddresses.Values()

	params, err := script.Code.Parameter.ToTypedAST()
	if err != nil {
		return err
	}
	contractScript.Entrypoints = params.GetEntrypoints()

	if script.IsUpgradable() {
		contractScript.Tags.Set(types.UpgradableTag)
	}

	proto, err := p.ctx.CachedProtocolByID(operation.Network, operation.ProtocolID)
	if err != nil {
		return err
	}

	if contractScript.ID > 0 {
		c.AlphaID = contractScript.ID
		switch proto.SymLink {
		case bcd.SymLinkAlpha:
		case bcd.SymLinkBabylon:
			c.BabylonID = c.AlphaID
		}
	} else {
		c.Alpha = contractScript
		switch proto.SymLink {
		case bcd.SymLinkAlpha:
		case bcd.SymLinkBabylon:
			c.Babylon = c.Alpha
		}
	}

	c.Tags = contractScript.Tags

	return nil
}

func (p *Parser) replaceConstants(c *contract.Script, operation *operation.Operation) {
	pattern := `{"prim":"constant","args":[{"string":"%s"}]}`
	for i := range c.Constants {
		operation.Script = bytes.ReplaceAll(
			operation.Script,
			[]byte(fmt.Sprintf(pattern, c.Constants[i].Address)),
			c.Constants[i].Value,
		)
	}
}
