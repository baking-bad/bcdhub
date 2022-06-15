package contract

import (
	"encoding/json"

	"github.com/baking-bad/bcdhub/internal/bcd"
	astContract "github.com/baking-bad/bcdhub/internal/bcd/contract"
	"github.com/baking-bad/bcdhub/internal/config"
	"github.com/baking-bad/bcdhub/internal/models/contract"
	"github.com/baking-bad/bcdhub/internal/models/operation"
	"github.com/baking-bad/bcdhub/internal/models/types"
	"github.com/baking-bad/bcdhub/internal/parsers"
	"github.com/pkg/errors"
)

// Hangzhou -
type Hangzhou struct {
	ctx *config.Context
}

// NewHangzhou -
func NewHangzhou(ctx *config.Context) *Hangzhou {
	return &Hangzhou{ctx: ctx}
}

// Parse -
func (p *Hangzhou) Parse(operation *operation.Operation, store parsers.Store) error {
	if !operation.IsOrigination() {
		return errors.Errorf("invalid operation kind in computeContractMetrics: %s", operation.Kind)
	}

	contract := contract.Contract{
		Level:      operation.Level,
		Timestamp:  operation.Timestamp,
		Manager:    operation.Source,
		Account:    operation.Destination,
		Delegate:   operation.Delegate,
		LastAction: operation.Timestamp,
	}

	if err := p.computeMetrics(operation, &contract); err != nil {
		return err
	}

	store.AddContracts(&contract)
	return nil
}

func (p *Hangzhou) computeMetrics(operation *operation.Operation, c *contract.Contract) error {
	script, err := astContract.NewParser(operation.Script)
	if err != nil {
		return errors.Wrap(err, "astContract.NewParser")
	}
	operation.AST = script.Code

	contractScript, err := p.ctx.Scripts.ByHash(script.Hash)
	if err != nil {
		if !p.ctx.Storage.IsRecordNotFound(err) {
			return err
		}
		var s bcd.RawScript
		if err := json.Unmarshal(script.CodeRaw, &s); err != nil {
			return err
		}

		constants, err := script.FindConstants()
		if err != nil {
			return errors.Wrap(err, "script.FindConstants")
		}

		if len(constants) > 0 {
			globalConstants, err := p.ctx.GlobalConstants.All(constants...)
			if err != nil {
				return err
			}
			contractScript.Constants = globalConstants
			replaceConstants(&contractScript, operation)

			script, err = astContract.NewParser(operation.Script)
			if err != nil {
				return errors.Wrap(err, "astContract.NewParser")
			}
			operation.AST = script.Code
		}

		if err := script.Parse(); err != nil {
			return err
		}

		params, err := script.Code.Parameter.ToTypedAST()
		if err != nil {
			return err
		}

		operation.Script = script.CodeRaw
		contractScript.Hash = script.Hash
		contractScript.Code = s.Code
		contractScript.Parameter = s.Parameter
		contractScript.Storage = s.Storage
		contractScript.Views = s.Views
		contractScript.FailStrings = script.FailStrings.Values()
		contractScript.Annotations = script.Annotations.Values()
		contractScript.Tags = types.NewTags(script.Tags.Values())
		contractScript.Hardcoded = script.HardcodedAddresses.Values()
		contractScript.Entrypoints = params.GetEntrypoints()

		c.Babylon = contractScript
	} else {
		c.BabylonID = contractScript.ID
		c.Babylon = contractScript
	}

	c.Tags = contractScript.Tags

	return nil
}
