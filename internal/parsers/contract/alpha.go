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

// Alpha -
type Alpha struct {
	ctx *config.Context
}

// NewAlpha -
func NewAlpha(ctx *config.Context) *Alpha {
	return &Alpha{ctx: ctx}
}

// Parse -
func (p *Alpha) Parse(ctx context.Context, operation *operation.Operation, store parsers.Store) error {
	if !operation.IsOrigination() {
		return errors.Errorf("invalid operation kind in alpha.parse: %s", operation.Kind)
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

func (p *Alpha) computeMetrics(ctx context.Context, operation *operation.Operation, c *contract.Contract) error {
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
		contractScript.Level = operation.Level

		c.Alpha = contractScript
	} else {
		c.AlphaID = contractScript.ID
		c.Alpha = contractScript
	}

	c.Tags = contractScript.Tags

	return nil
}
