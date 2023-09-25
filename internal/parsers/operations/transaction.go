package operations

import (
	"context"

	"github.com/baking-bad/bcdhub/internal/bcd"
	"github.com/baking-bad/bcdhub/internal/bcd/ast"
	"github.com/baking-bad/bcdhub/internal/bcd/consts"
	"github.com/baking-bad/bcdhub/internal/bcd/types"
	"github.com/baking-bad/bcdhub/internal/models/account"
	"github.com/baking-bad/bcdhub/internal/models/operation"
	modelsTypes "github.com/baking-bad/bcdhub/internal/models/types"
	"github.com/baking-bad/bcdhub/internal/noderpc"
	"github.com/baking-bad/bcdhub/internal/parsers"
	"github.com/pkg/errors"

	jsoniter "github.com/json-iterator/go"
)

var json = jsoniter.ConfigCompatibleWithStandardLibrary

// Transaction -
type Transaction struct {
	*ParseParams
}

// NewTransaction -
func NewTransaction(params *ParseParams) Transaction {
	return Transaction{params}
}

// Parse -
func (p Transaction) Parse(ctx context.Context, data noderpc.Operation, store parsers.Store) error {
	source := account.Account{
		Address: data.Source,
		Type:    modelsTypes.NewAccountType(data.Source),
		Level:   p.head.Level,
	}

	tx := operation.Operation{
		Hash:         p.hash,
		ProtocolID:   p.protocol.ID,
		Level:        p.head.Level,
		Timestamp:    p.head.Timestamp,
		Kind:         modelsTypes.NewOperationKind(data.Kind),
		Initiator:    source,
		Source:       source,
		Fee:          data.Fee,
		Counter:      data.Counter,
		GasLimit:     data.GasLimit,
		StorageLimit: data.StorageLimit,
		Amount:       *data.Amount,
		Destination: account.Account{
			Address: *data.Destination,
			Type:    modelsTypes.NewAccountType(*data.Destination),
			Level:   p.head.Level,
		},
		Delegate: account.Account{
			Address: data.Delegate,
			Type:    modelsTypes.NewAccountType(data.Delegate),
			Level:   p.head.Level,
		},
		Nonce:        data.Nonce,
		Parameters:   data.Parameters,
		ContentIndex: p.contentIdx,
	}

	p.fillInternal(&tx)

	parseOperationResult(data, &tx, store)

	tx.SetBurned(*p.protocol.Constants)

	store.AddOperations(&tx)
	store.AddAccounts(
		&tx.Source,
		&tx.Delegate,
		&tx.Destination,
	)

	switch tx.Destination.Type {
	case modelsTypes.AccountTypeContract:
		return p.parseContractParams(ctx, data, store, &tx)
	case modelsTypes.AccountTypeSmartRollup:
		return p.parseSmartRollupParams(data, &tx)
	default:
		return nil
	}
}

func (p Transaction) parseSmartRollupParams(data noderpc.Operation, tx *operation.Operation) error {
	if len(tx.Parameters) == 0 {
		return tx.Entrypoint.Set(consts.DefaultEntrypoint)
	}

	params := types.NewParameters(tx.Parameters)
	if err := tx.Entrypoint.Set(params.Entrypoint); err != nil {
		return err
	}
	return nil
}

func (p Transaction) parseContractParams(ctx context.Context, data noderpc.Operation, store parsers.Store, tx *operation.Operation) error {
	for i := range tx.Errors {
		if tx.Errors[i].Is("contract.non_existing_contract") {
			return nil
		}
	}

	scriptEntity, err := p.ctx.Contracts.Script(ctx, tx.Destination.Address, p.protocol.SymLink)
	if err != nil {
		if !tx.Internal {
			return nil
		}

		contracts := store.ListContracts()
		for i := range contracts {
			if tx.Destination.Address == contracts[i].Account.Address {
				switch p.protocol.SymLink {
				case bcd.SymLinkAlpha:
					tx.Script, err = contracts[i].Alpha.Full()
					if err != nil {
						return err
					}
				case bcd.SymLinkBabylon:
					tx.Script, err = contracts[i].Babylon.Full()
					if err != nil {
						return err
					}
				case bcd.SymLinkJakarta:
					tx.Script, err = contracts[i].Jakarta.Full()
					if err != nil {
						return err
					}
				default:
					return errors.Errorf("unknown protocol symbolic link: %s", p.protocol.SymLink)
				}
			}
		}
		if tx.Script == nil {
			return err
		}
	} else {
		tx.Script, err = scriptEntity.Full()
		if err != nil {
			return err
		}
	}

	tx.AST, err = ast.NewScriptWithoutCode(tx.Script)
	if err != nil {
		return err
	}

	if err := setTags(ctx, p.ctx, nil, tx); err != nil {
		return err
	}

	if err := p.getEntrypoint(tx); err != nil {
		return err
	}
	p.stackTrace.Add(*tx)

	if tx.IsApplied() {
		if err := p.appliedHandler(ctx, data, tx, store); err != nil {
			return err
		}
	}
	return nil
}

func (p Transaction) fillInternal(tx *operation.Operation) {
	if p.main == nil {
		p.main = tx
		return
	}

	tx.Counter = p.main.Counter
	tx.Hash = p.main.Hash
	tx.Level = p.main.Level
	tx.Timestamp = p.main.Timestamp
	tx.Internal = true
	tx.Initiator = p.main.Source
}

func (p Transaction) appliedHandler(ctx context.Context, item noderpc.Operation, tx *operation.Operation, store parsers.Store) error {
	if err := p.specific.StorageParser.ParseTransaction(ctx, item, tx, store); err != nil {
		return err
	}

	return NewMigration(p.ctx.Contracts).Parse(ctx, item, tx, p.protocol.Hash, store)
}

func (p Transaction) getEntrypoint(tx *operation.Operation) error {
	if len(tx.Parameters) == 0 {
		return tx.Entrypoint.Set(consts.DefaultEntrypoint)
	}

	params := types.NewParameters(tx.Parameters)
	if err := tx.Entrypoint.Set(params.Entrypoint); err != nil {
		return err
	}

	if !tx.IsApplied() {
		return nil
	}

	param, err := tx.AST.ParameterType()
	if err != nil {
		return err
	}

	subTree, err := param.FromParameters(params)
	if err != nil {
		return err
	}

	node, entrypointName := subTree.UnwrapAndGetEntrypointName()
	if node == nil {
		return nil
	}

	return tx.Entrypoint.Set(entrypointName)
}
