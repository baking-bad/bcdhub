package postgres

import (
	"context"

	"github.com/baking-bad/bcdhub/internal/models"
	"github.com/baking-bad/bcdhub/internal/models/bigmapdiff"
	"github.com/baking-bad/bcdhub/internal/models/contract"
	"github.com/baking-bad/bcdhub/internal/models/migration"
	"github.com/baking-bad/bcdhub/internal/models/operation"
	smartrollup "github.com/baking-bad/bcdhub/internal/models/smart_rollup"
	"github.com/baking-bad/bcdhub/internal/models/types"
	"github.com/pkg/errors"
)

// Store -
type Store struct {
	BigMapState     []*bigmapdiff.BigMapState
	Contracts       []*contract.Contract
	Migrations      []*migration.Migration
	Operations      []*operation.Operation
	GlobalConstants []*contract.GlobalConstant
	SmartRollups    []*smartrollup.SmartRollup

	partitions *PartitionManager
}

// NewStore -
func NewStore(pm *PartitionManager) *Store {
	return &Store{
		BigMapState:     make([]*bigmapdiff.BigMapState, 0),
		Contracts:       make([]*contract.Contract, 0),
		Migrations:      make([]*migration.Migration, 0),
		Operations:      make([]*operation.Operation, 0),
		GlobalConstants: make([]*contract.GlobalConstant, 0),
		SmartRollups:    make([]*smartrollup.SmartRollup, 0),
		partitions:      pm,
	}
}

// AddBigMapStates -
func (store *Store) AddBigMapStates(states ...*bigmapdiff.BigMapState) {
	store.BigMapState = append(store.BigMapState, states...)
}

// AddContracts -
func (store *Store) AddContracts(contracts ...*contract.Contract) {
	store.Contracts = append(store.Contracts, contracts...)
}

// AddMigrations -
func (store *Store) AddMigrations(migrations ...*migration.Migration) {
	store.Migrations = append(store.Migrations, migrations...)
}

// AddOperations -
func (store *Store) AddOperations(operations ...*operation.Operation) {
	store.Operations = append(store.Operations, operations...)
}

// AddGlobalConstants -
func (store *Store) AddGlobalConstants(constants ...*contract.GlobalConstant) {
	store.GlobalConstants = append(store.GlobalConstants, constants...)
}

// AddSmartRollups -
func (store *Store) AddSmartRollups(rollups ...*smartrollup.SmartRollup) {
	store.SmartRollups = append(store.SmartRollups, rollups...)
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
func (store *Store) Save(ctx context.Context, tx models.Transaction) error {
	if err := store.saveOperations(ctx, tx); err != nil {
		return err
	}

	if err := store.saveContracts(ctx, tx); err != nil {
		return err
	}

	if err := store.saveMigrations(ctx, tx); err != nil {
		return err
	}

	if err := tx.BigMapStates(ctx, store.BigMapState...); err != nil {
		return errors.Wrap(err, "saving bigmap states")
	}

	if err := tx.GlobalConstants(ctx, store.GlobalConstants...); err != nil {
		return errors.Wrap(err, "saving bigmap states")
	}

	if err := store.saveSmartRollups(ctx, tx); err != nil {
		return err
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

	for i := range store.SmartRollups {
		if !store.SmartRollups[i].Address.IsEmpty() {
			if err := tx.Accounts(ctx, &store.SmartRollups[i].Address); err != nil {
				return err
			}
			store.SmartRollups[i].AddressId = store.SmartRollups[i].Address.ID
		}
	}

	return tx.SmartRollups(ctx, store.SmartRollups...)
}

func (store *Store) saveOperations(ctx context.Context, tx models.Transaction) error {
	if len(store.Operations) == 0 {
		return nil
	}

	if err := store.partitions.CreatePartitions(ctx, store.Operations[0].Timestamp); err != nil {
		return err
	}

	for i := range store.Operations {
		if !store.Operations[i].Source.IsEmpty() {
			if err := tx.Accounts(ctx, &store.Operations[i].Source); err != nil {
				return err
			}
			store.Operations[i].SourceID = store.Operations[i].Source.ID
		}
		if !store.Operations[i].Destination.IsEmpty() {
			if err := tx.Accounts(ctx, &store.Operations[i].Destination); err != nil {
				return err
			}
			store.Operations[i].DestinationID = store.Operations[i].Destination.ID
		}
		if !store.Operations[i].Initiator.IsEmpty() {
			if err := tx.Accounts(ctx, &store.Operations[i].Initiator); err != nil {
				return err
			}
			store.Operations[i].InitiatorID = store.Operations[i].Initiator.ID
		}
		if !store.Operations[i].Delegate.IsEmpty() {
			if err := tx.Accounts(ctx, &store.Operations[i].Delegate); err != nil {
				return err
			}
			store.Operations[i].DelegateID = store.Operations[i].Delegate.ID
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
		for j := range store.Operations[i].TickerUpdates {
			if !store.Operations[i].TickerUpdates[j].Account.IsEmpty() {
				if err := tx.Accounts(ctx, &store.Operations[i].TickerUpdates[j].Account); err != nil {
					return err
				}
				store.Operations[i].TickerUpdates[j].AccountID = store.Operations[i].TickerUpdates[j].Account.ID
			}
			if !store.Operations[i].TickerUpdates[j].Ticketer.IsEmpty() {
				if err := tx.Accounts(ctx, &store.Operations[i].TickerUpdates[j].Ticketer); err != nil {
					return err
				}
				store.Operations[i].TickerUpdates[j].TicketerID = store.Operations[i].TickerUpdates[j].Ticketer.ID
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
	return store.updateContracts(ctx, tx)
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

		if err := tx.Accounts(ctx, &store.Contracts[i].Account); err != nil {
			return err
		}
		store.Contracts[i].AccountID = store.Contracts[i].Account.ID

		if !store.Contracts[i].Manager.IsEmpty() {
			if err := tx.Accounts(ctx, &store.Contracts[i].Manager); err != nil {
				return err
			}
			store.Contracts[i].ManagerID = store.Contracts[i].Manager.ID
		}
		if !store.Contracts[i].Delegate.IsEmpty() {
			if err := tx.Accounts(ctx, &store.Contracts[i].Delegate); err != nil {
				return err
			}
			store.Contracts[i].DelegateID = store.Contracts[i].Delegate.ID
		}
	}

	if err := tx.Contracts(ctx, store.Contracts...); err != nil {
		return err
	}

	return store.updateContracts(ctx, tx)
}

func (store *Store) updateContracts(ctx context.Context, tx models.Transaction) error {
	if len(store.Operations) == 0 {
		return nil
	}
	count := make(map[int64]uint64)
	for i := range store.Operations {
		destination := store.Operations[i].Destination
		if destination.Type != types.AccountTypeContract {
			continue
		}

		if value, ok := count[destination.ID]; ok {
			count[destination.ID] = value + 1
		} else {
			count[destination.ID] = 1
		}

		source := store.Operations[i].Source
		if source.Type != types.AccountTypeContract {
			continue
		}

		if value, ok := count[source.ID]; ok {
			count[source.ID] = value + 1
		} else {
			count[source.ID] = 1
		}
	}

	if len(count) == 0 {
		return nil
	}

	contracts := make([]*contract.Update, 0, len(count))
	for accountID, txCount := range count {
		contracts = append(contracts, &contract.Update{
			LastAction: store.Operations[0].Timestamp,
			AccountID:  accountID,
			TxCount:    txCount,
		})
	}
	return tx.UpdateContracts(ctx, contracts...)
}
