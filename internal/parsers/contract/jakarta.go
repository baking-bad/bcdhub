package contract

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"

	"github.com/baking-bad/bcdhub/internal/bcd"
	"github.com/baking-bad/bcdhub/internal/bcd/ast"
	astContract "github.com/baking-bad/bcdhub/internal/bcd/contract"
	bcdTypes "github.com/baking-bad/bcdhub/internal/bcd/types"
	"github.com/baking-bad/bcdhub/internal/config"
	"github.com/baking-bad/bcdhub/internal/models/contract"
	"github.com/baking-bad/bcdhub/internal/models/operation"
	"github.com/baking-bad/bcdhub/internal/models/types"
	"github.com/baking-bad/bcdhub/internal/parsers"
	"github.com/pkg/errors"
)

// Jakarta -
type Jakarta struct {
	ctx *config.Context
}

// NewJakarta -
func NewJakarta(ctx *config.Context) *Jakarta {
	return &Jakarta{ctx: ctx}
}

// Parse -
func (p *Jakarta) Parse(ctx context.Context, operation *operation.Operation, store parsers.Store) error {
	if !operation.IsOrigination() {
		return errors.Errorf("invalid operation kind in jakarta.parse: %s", operation.Kind)
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
	store.AddAccounts(&contract.Account)
	return nil
}

func (p *Jakarta) computeMetrics(ctx context.Context, operation *operation.Operation, c *contract.Contract) error {
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

		c.Jakarta = contractScript
	} else {
		c.JakartaID = contractScript.ID
		c.Jakarta = contractScript
	}

	c.Tags = contractScript.Tags

	return nil
}

func replaceConstants(constants []contract.GlobalConstant, operation *operation.Operation) {
	pattern := `{"prim":"constant","args":[{"string":"%s"}]}`
	for i := range constants {
		operation.Script = bytes.ReplaceAll(
			operation.Script,
			[]byte(fmt.Sprintf(pattern, constants[i].Address)),
			constants[i].Value,
		)
	}
}

func getGlobalConstants(ctx context.Context, repo contract.ConstantRepository, operation *operation.Operation) ([]contract.GlobalConstant, error) {
	var cd astContract.ContractData
	if err := json.Unmarshal(operation.Script, &cd); err != nil {
		return nil, err
	}

	var tree ast.UntypedAST
	if err := json.Unmarshal(cd.Code, &tree); err != nil {
		return nil, err
	}

	constants, err := astContract.FindConstants(tree)
	if err != nil {
		return nil, err
	}

	globalConstants := make(bcdTypes.Set)
	globalContantsModels := make([]contract.GlobalConstant, 0)

	for len(constants) > 0 {
		addresses := make([]string, 0)
		for address := range constants {
			if !globalConstants.Has(address) {
				addresses = append(addresses, address)
				globalConstants.Add(address)
			}
		}

		entities, err := repo.All(ctx, addresses...)
		if err != nil {
			return nil, err
		}
		globalContantsModels = append(globalContantsModels, entities...)
		replaceConstants(entities, operation)

		var tree ast.UntypedAST
		if err := json.Unmarshal(operation.Script, &tree); err != nil {
			return nil, err
		}
		constants, err = astContract.FindConstants(tree)
		if err != nil {
			return nil, err
		}
	}

	return globalContantsModels, nil
}
