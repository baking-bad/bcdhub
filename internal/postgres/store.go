package postgres

import (
	"context"

	"github.com/baking-bad/bcdhub/internal/models"
	"github.com/baking-bad/bcdhub/internal/models/account"
	"github.com/baking-bad/bcdhub/internal/models/bigmapdiff"
	"github.com/baking-bad/bcdhub/internal/models/block"
	"github.com/baking-bad/bcdhub/internal/models/contract"
	"github.com/baking-bad/bcdhub/internal/models/migration"
	"github.com/baking-bad/bcdhub/internal/models/operation"
	smartrollup "github.com/baking-bad/bcdhub/internal/models/smart_rollup"
	"github.com/baking-bad/bcdhub/internal/models/stats"
	"github.com/baking-bad/bcdhub/internal/models/types"
	"github.com/baking-bad/bcdhub/internal/postgres/core"
	"github.com/pkg/errors"
	"github.com/uptrace/bun"
)

// Store -
type Store struct {
	Block           *block.Block
	BigMapState     []*bigmapdiff.BigMapState
	Contracts       []*contract.Contract
	Migrations      []*migration.Migration
	Operations      []*operation.Operation
	GlobalConstants []*contract.GlobalConstant
	SmartRollups    []*smartrollup.SmartRollup
	Accounts        map[string]*account.Account
	Stats           stats.Stats

	stats  stats.Repository
	db     *bun.DB
	accIds map[string]int64
}

// NewStore -
func NewStore(db *bun.DB, statsRepo stats.Repository) *Store {
	return &Store{
		BigMapState:     make([]*bigmapdiff.BigMapState, 0),
		Contracts:       make([]*contract.Contract, 0),
		Migrations:      make([]*migration.Migration, 0),
		Operations:      make([]*operation.Operation, 0),
		GlobalConstants: make([]*contract.GlobalConstant, 0),
		SmartRollups:    make([]*smartrollup.SmartRollup, 0),
		Accounts:        make(map[string]*account.Account),
		Stats:           stats.Stats{},
		stats:           statsRepo,
		db:              db,
		accIds:          make(map[string]int64),
	}
}

func (store *Store) SetBlock(block *block.Block) {
	store.Block = block
}

// AddBigMapStates -
func (store *Store) AddBigMapStates(states ...*bigmapdiff.BigMapState) {
	store.BigMapState = append(store.BigMapState, states...)
}

// AddContracts -
func (store *Store) AddContracts(contracts ...*contract.Contract) {
	store.Contracts = append(store.Contracts, contracts...)

	store.Stats.ContractsCount += len(contracts)
}

// AddMigrations -
func (store *Store) AddMigrations(migrations ...*migration.Migration) {
	store.Migrations = append(store.Migrations, migrations...)
}

// AddOperations -
func (store *Store) AddOperations(operations ...*operation.Operation) {
	store.Operations = append(store.Operations, operations...)

	store.Stats.OperationsCount += len(operations)
	for i := range operations {
		switch operations[i].Kind {
		case types.OperationKindEvent:
			store.Stats.EventsCount += 1
		case types.OperationKindOrigination, types.OperationKindOriginationNew:
			store.Stats.OriginationsCount += 1
		case types.OperationKindSrOrigination:
			store.Stats.SrOriginationsCount += 1
		case types.OperationKindTransaction:
			store.Stats.TransactionsCount += 1
		case types.OperationKindRegisterGlobalConstant:
			store.Stats.RegisterGlobalConstantCount += 1
		case types.OperationKindSrExecuteOutboxMessage:
			store.Stats.SrExecutesCount += 1
		}
	}
}

// AddGlobalConstants -
func (store *Store) AddGlobalConstants(constants ...*contract.GlobalConstant) {
	store.GlobalConstants = append(store.GlobalConstants, constants...)
	store.Stats.GlobalConstantsCount += len(constants)
}

// AddSmartRollups -
func (store *Store) AddSmartRollups(rollups ...*smartrollup.SmartRollup) {
	store.SmartRollups = append(store.SmartRollups, rollups...)
	store.Stats.ContractsCount += len(rollups)
}

// AddAccounts -
func (store *Store) AddAccounts(accounts ...*account.Account) {
	for i := range accounts {
		if account, ok := store.Accounts[accounts[i].Address]; !ok {
			store.Accounts[accounts[i].Address] = accounts[i]
		} else {
			account.OperationsCount += accounts[i].OperationsCount
		}
	}
}

// ListContracts -
func (store *Store) ListContracts() []*contract.Contract {
	return store.Contracts
}

// ListOperations -
func (store *Store) ListOperations() []*operation.Operation {
	return store.Operations
}

// Save -
func (store *Store) Save(ctx context.Context) error {
	stats, err := store.stats.Get(ctx)
	if err != nil {
		return err
	}
	store.Stats.ID = stats.ID

	tx, err := core.NewTransaction(ctx, store.db)
	if err != nil {
		return err
	}

	if err := tx.Block(ctx, store.Block); err != nil {
		return errors.Wrap(err, "saving block")
	}

	if err := store.saveAccounts(ctx, tx); err != nil {
		return errors.Wrap(err, "saving accounts")
	}

	if err := store.saveOperations(ctx, tx); err != nil {
		return errors.Wrap(err, "saving operations")
	}

	if err := store.saveContracts(ctx, tx); err != nil {
		return errors.Wrap(err, "saving contracts")
	}

	if err := store.saveMigrations(ctx, tx); err != nil {
		return errors.Wrap(err, "saving migrations")
	}

	if err := tx.BigMapStates(ctx, store.BigMapState...); err != nil {
		return errors.Wrap(err, "saving bigmap states")
	}

	if err := tx.GlobalConstants(ctx, store.GlobalConstants...); err != nil {
		return errors.Wrap(err, "saving bigmap states")
	}

	if err := store.saveSmartRollups(ctx, tx); err != nil {
		return errors.Wrap(err, "saving smart rollups")
	}

	if err := tx.UpdateStats(ctx, store.Stats); err != nil {
		return errors.Wrap(err, "saving stats")
	}

	return tx.Commit()
}

func (store *Store) saveAccounts(ctx context.Context, tx models.Transaction) error {
	if len(store.Accounts) == 0 {
		return nil
	}

	arr := make([]*account.Account, 0, len(store.Accounts))
	for _, acc := range store.Accounts {
		if acc.IsEmpty() {
			continue
		}
		arr = append(arr, acc)
	}

	if err := tx.Accounts(ctx, arr...); err != nil {
		return err
	}

	for i := range arr {
		store.accIds[arr[i].Address] = arr[i].ID
	}

	return nil
}

func (store *Store) saveMigrations(ctx context.Context, tx models.Transaction) error {
	if len(store.Migrations) == 0 {
		return nil
	}

	for i := range store.Migrations {
		if store.Migrations[i].ContractID == 0 {
			store.Migrations[i].ContractID = store.Migrations[i].Contract.ID
		}
	}

	return tx.Migrations(ctx, store.Migrations...)
}

func (store *Store) saveSmartRollups(ctx context.Context, tx models.Transaction) error {
	if len(store.SmartRollups) == 0 {
		return nil
	}

	for i, rollup := range store.SmartRollups {
		if !rollup.Address.IsEmpty() {
			if id, ok := store.accIds[rollup.Address.Address]; ok {
				store.SmartRollups[i].AddressId = id
			} else {
				return errors.Errorf("unknown smart rollup account: %s", rollup.Address.Address)
			}
		}
	}

	return tx.SmartRollups(ctx, store.SmartRollups...)
}

func (store *Store) saveOperations(ctx context.Context, tx models.Transaction) error {
	if len(store.Operations) == 0 {
		return nil
	}

	for i, operation := range store.Operations {
		if !operation.Source.IsEmpty() {
			if id, ok := store.accIds[operation.Source.Address]; ok {
				store.Operations[i].SourceID = id
			} else {
				return errors.Errorf("unknown source account: %s", operation.Source.Address)
			}
		}
		if !operation.Destination.IsEmpty() {
			if id, ok := store.accIds[operation.Destination.Address]; ok {
				store.Operations[i].DestinationID = id
			} else {
				return errors.Errorf("unknown destination account: %s", operation.Destination.Address)
			}
		}
		if !store.Operations[i].Initiator.IsEmpty() {
			if id, ok := store.accIds[operation.Initiator.Address]; ok {
				store.Operations[i].InitiatorID = id
			} else {
				return errors.Errorf("unknown initiator account: %s", operation.Initiator.Address)
			}
		}
		if !store.Operations[i].Delegate.IsEmpty() {
			if id, ok := store.accIds[operation.Delegate.Address]; ok {
				store.Operations[i].DelegateID = id
			} else {
				return errors.Errorf("unknown delegate account: %s", operation.Delegate.Address)
			}
		}
	}

	if err := tx.Operations(ctx, store.Operations...); err != nil {
		return errors.Wrap(err, "saving operations")
	}

	for i := range store.Operations {
		for j := range store.Operations[i].BigMapDiffs {
			store.Operations[i].BigMapDiffs[j].OperationID = store.Operations[i].ID
		}
		for j := range store.Operations[i].BigMapActions {
			store.Operations[i].BigMapActions[j].OperationID = store.Operations[i].ID
		}
		for j, update := range store.Operations[i].TickerUpdates {
			if !update.Account.IsEmpty() {
				if id, ok := store.accIds[update.Account.Address]; ok {
					store.Operations[i].TickerUpdates[j].AccountID = id
				} else {
					return errors.Errorf("unknown ticket update account: %s", update.Account.Address)
				}
			}
			if !update.Ticketer.IsEmpty() {
				if id, ok := store.accIds[update.Ticketer.Address]; ok {
					store.Operations[i].TickerUpdates[j].TicketerID = id
				} else {
					return errors.Errorf("unknown ticket update ticketer account: %s", update.Ticketer.Address)
				}
			}
			store.Operations[i].TickerUpdates[j].OperationID = store.Operations[i].ID
		}

		if err := tx.BigMapDiffs(ctx, store.Operations[i].BigMapDiffs...); err != nil {
			return errors.Wrap(err, "saving bigmap diffs")
		}
		if err := tx.BigMapActions(ctx, store.Operations[i].BigMapActions...); err != nil {
			return errors.Wrap(err, "saving bigmap actions")
		}
		if err := tx.TickerUpdates(ctx, store.Operations[i].TickerUpdates...); err != nil {
			return errors.Wrap(err, "saving ticket updates")
		}
	}
	return nil
}

func (store *Store) saveContracts(ctx context.Context, tx models.Transaction) error {
	if len(store.Contracts) == 0 {
		return nil
	}

	for i := range store.Contracts {
		if store.Contracts[i].Alpha.Code != nil {
			if err := tx.Scripts(ctx, &store.Contracts[i].Alpha); err != nil {
				return err
			}
			store.Contracts[i].AlphaID = store.Contracts[i].Alpha.ID
		}
		if store.Contracts[i].Babylon.Code != nil {
			if store.Contracts[i].Alpha.Hash != store.Contracts[i].Babylon.Hash {
				if err := tx.Scripts(ctx, &store.Contracts[i].Babylon); err != nil {
					return err
				}
				store.Contracts[i].BabylonID = store.Contracts[i].Babylon.ID

				if len(store.Contracts[i].Babylon.Constants) > 0 {
					for j := range store.Contracts[i].Babylon.Constants {
						relation := contract.ScriptConstants{
							ScriptId:         store.Contracts[i].BabylonID,
							GlobalConstantId: store.Contracts[i].Babylon.Constants[j].ID,
						}
						if err := tx.ScriptConstant(ctx, &relation); err != nil {
							return err
						}
					}
				}

			} else {
				store.Contracts[i].BabylonID = store.Contracts[i].Alpha.ID
			}
		}
		if store.Contracts[i].Jakarta.Code != nil {
			if store.Contracts[i].Babylon.Hash != store.Contracts[i].Jakarta.Hash {
				if err := tx.Scripts(ctx, &store.Contracts[i].Jakarta); err != nil {
					return err
				}
				store.Contracts[i].JakartaID = store.Contracts[i].Jakarta.ID

				if len(store.Contracts[i].Jakarta.Constants) > 0 {
					for j := range store.Contracts[i].Jakarta.Constants {
						relation := contract.ScriptConstants{
							ScriptId:         store.Contracts[i].JakartaID,
							GlobalConstantId: store.Contracts[i].Jakarta.Constants[j].ID,
						}
						if err := tx.ScriptConstant(ctx, &relation); err != nil {
							return err
						}
					}
				}

			} else {
				store.Contracts[i].JakartaID = store.Contracts[i].Babylon.ID
			}
		}

		if id, ok := store.accIds[store.Contracts[i].Account.Address]; ok {
			store.Contracts[i].AccountID = id
		} else {
			return errors.Errorf("unknown contract account: %s", store.Contracts[i].Account.Address)
		}

		if !store.Contracts[i].Manager.IsEmpty() {
			if id, ok := store.accIds[store.Contracts[i].Manager.Address]; ok {
				store.Contracts[i].ManagerID = id
			} else {
				return errors.Errorf("unknown manager account: %s", store.Contracts[i].Manager.Address)
			}
		}
		if !store.Contracts[i].Delegate.IsEmpty() {
			if id, ok := store.accIds[store.Contracts[i].Delegate.Address]; ok {
				store.Contracts[i].DelegateID = id
			} else {
				return errors.Errorf("unknown delegate account: %s", store.Contracts[i].Delegate.Address)
			}
		}
	}

	if err := tx.Contracts(ctx, store.Contracts...); err != nil {
		return err
	}

	return nil
}
