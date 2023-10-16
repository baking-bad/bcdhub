package contract

import (
	"context"
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
func (p *Hangzhou) Parse(ctx context.Context, operation *operation.Operation, store parsers.Store) error {
	if !operation.IsOrigination() {
		return errors.Errorf("invalid operation kind in hangzhou.parse: %s", operation.Kind)
	}

	contract := contract.Contract{
		Level:     operation.Level,
		Timestamp: operation.Timestamp,
		Manager:   operation.Source,
		Account:   operation.Destination,
		Delegate:  operation.Delegate,
	}

	if err := p.computeMetrics(ctx, operation, &contract); err != nil {
		return err
	}

	store.AddContracts(&contract)
	return nil
}

func (p *Hangzhou) computeMetrics(ctx context.Context, operation *operation.Operation, c *contract.Contract) error {
	constants, err := getGlobalConstants(ctx, p.ctx.GlobalConstants, operation)
	if err != nil {
		return errors.Wrap(err, "getGlobalConstants")
	}

	script, err := astContract.NewParser(operation.Script)
	if err != nil {
		return errors.Wrap(err, "astContract.NewParser")
	}
	operation.AST = script.Code

	contractScript, err := p.ctx.Scripts.ByHash(ctx, script.Hash)
	if err != nil {
		if !p.ctx.Storage.IsRecordNotFound(err) {
			return err
		}
		var s bcd.RawScript
		if err := json.Unmarshal(script.CodeRaw, &s); err != nil {
			return err
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
		contractScript.Constants = constants
		contractScript.Level = operation.Level

		c.Babylon = contractScript
	} else {
		c.BabylonID = contractScript.ID
		c.Babylon = contractScript
	}

	c.Tags = contractScript.Tags

	return nil
}
