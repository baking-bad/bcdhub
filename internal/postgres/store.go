package postgres

import (
	"context"
	"time"

	"github.com/baking-bad/bcdhub/internal/models/bigmapdiff"
	"github.com/baking-bad/bcdhub/internal/models/contract"
	"github.com/baking-bad/bcdhub/internal/models/migration"
	"github.com/baking-bad/bcdhub/internal/models/operation"
	smartrollup "github.com/baking-bad/bcdhub/internal/models/smart_rollup"
	"github.com/baking-bad/bcdhub/internal/models/types"
	"github.com/uptrace/bun"
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
	tx         bun.IDB
}

// NewStore -
func NewStore(tx bun.IDB, pm *PartitionManager) *Store {
	return &Store{
		BigMapState:     make([]*bigmapdiff.BigMapState, 0),
		Contracts:       make([]*contract.Contract, 0),
		Migrations:      make([]*migration.Migration, 0),
		Operations:      make([]*operation.Operation, 0),
		GlobalConstants: make([]*contract.GlobalConstant, 0),
		SmartRollups:    make([]*smartrollup.SmartRollup, 0),

		partitions: pm,
		tx:         tx,
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
func (store *Store) Save(ctx context.Context) error {
	if err := store.saveOperations(ctx, store.tx); err != nil {
		return err
	}

	if err := store.saveContracts(ctx, store.tx); err != nil {
		return err
	}

	if err := store.saveMigrations(ctx, store.tx); err != nil {
		return err
	}

	for i := range store.BigMapState {
		if err := store.BigMapState[i].Save(ctx, store.tx); err != nil {
			return err
		}
	}

	if len(store.GlobalConstants) > 0 {
		if _, err := store.tx.NewInsert().Model(&store.GlobalConstants).Returning("id").Exec(ctx); err != nil {
			return err
		}
	}

	if err := store.saveSmartRollups(ctx, store.tx); err != nil {
		return err
	}

	return nil
}

func (store *Store) saveMigrations(ctx context.Context, tx bun.IDB) error {
	if len(store.Migrations) == 0 {
		return nil
	}

	for i := range store.Migrations {
		if store.Migrations[i].ContractID == 0 {
			store.Migrations[i].ContractID = store.Migrations[i].Contract.ID
		}
	}

	_, err := tx.NewInsert().Model(&store.Migrations).Returning("id").Exec(ctx)
	return err
}

func (store *Store) saveSmartRollups(ctx context.Context, tx bun.IDB) error {
	if len(store.SmartRollups) == 0 {
		return nil
	}

	for i := range store.SmartRollups {
		if !store.SmartRollups[i].Address.IsEmpty() {
			if err := store.SmartRollups[i].Address.Save(ctx, tx); err != nil {
				return err
			}
			store.SmartRollups[i].AddressId = store.SmartRollups[i].Address.ID
		}
	}

	_, err := tx.NewInsert().Model(&store.SmartRollups).Returning("id").Exec(ctx)
	return err
}

func (store *Store) saveOperations(ctx context.Context, tx bun.IDB) error {
	if len(store.Operations) == 0 {
		return nil
	}

	for i := range store.Operations {
		if !store.Operations[i].Source.IsEmpty() {
			if err := store.Operations[i].Source.Save(ctx, tx); err != nil {
				return err
			}
			store.Operations[i].SourceID = store.Operations[i].Source.ID
		}
		if !store.Operations[i].Destination.IsEmpty() {
			if err := store.Operations[i].Destination.Save(ctx, tx); err != nil {
				return err
			}
			store.Operations[i].DestinationID = store.Operations[i].Destination.ID
		}
		if !store.Operations[i].Initiator.IsEmpty() {
			if err := store.Operations[i].Initiator.Save(ctx, tx); err != nil {
				return err
			}
			store.Operations[i].InitiatorID = store.Operations[i].Initiator.ID
		}
		if !store.Operations[i].Delegate.IsEmpty() {
			if err := store.Operations[i].Delegate.Save(ctx, tx); err != nil {
				return err
			}
			store.Operations[i].DelegateID = store.Operations[i].Delegate.ID
		}
	}

	if err := store.partitions.CreatePartitions(ctx, store.Operations[0].Timestamp); err != nil {
		return err
	}

	if _, err := tx.NewInsert().Model(&store.Operations).Returning("id").Exec(ctx); err != nil {
		return err
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
				if err := store.Operations[i].TickerUpdates[j].Account.Save(ctx, tx); err != nil {
					return err
				}
				store.Operations[i].TickerUpdates[j].AccountID = store.Operations[i].TickerUpdates[j].Account.ID
			}
			if !store.Operations[i].TickerUpdates[j].Ticketer.IsEmpty() {
				if err := store.Operations[i].TickerUpdates[j].Ticketer.Save(ctx, tx); err != nil {
					return err
				}
				store.Operations[i].TickerUpdates[j].TicketerID = store.Operations[i].TickerUpdates[j].Ticketer.ID
			}
			store.Operations[i].TickerUpdates[j].OperationID = store.Operations[i].ID
		}

		if len(store.Operations[i].BigMapDiffs) > 0 {
			if _, err := tx.NewInsert().Model(&store.Operations[i].BigMapDiffs).Returning("id").Exec(ctx); err != nil {
				return err
			}
		}

		if len(store.Operations[i].BigMapActions) > 0 {
			if _, err := tx.NewInsert().Model(&store.Operations[i].BigMapActions).Returning("id").Exec(ctx); err != nil {
				return err
			}
		}

		if len(store.Operations[i].TickerUpdates) > 0 {
			if _, err := tx.NewInsert().Model(&store.Operations[i].TickerUpdates).Returning("id").Exec(ctx); err != nil {
				return err
			}
		}
	}
	return store.updateContracts(ctx, tx)
}

func (store *Store) saveContracts(ctx context.Context, tx bun.IDB) error {
	if len(store.Contracts) == 0 {
		return nil
	}

	for i := range store.Contracts {
		if store.Contracts[i].Alpha.Code != nil {
			if err := store.Contracts[i].Alpha.Save(ctx, tx); err != nil {
				return err
			}
			store.Contracts[i].AlphaID = store.Contracts[i].Alpha.ID
		}
		if store.Contracts[i].Babylon.Code != nil {
			if store.Contracts[i].Alpha.Hash != store.Contracts[i].Babylon.Hash {
				if err := store.Contracts[i].Babylon.Save(ctx, tx); err != nil {
					return err
				}
				store.Contracts[i].BabylonID = store.Contracts[i].Babylon.ID

				if len(store.Contracts[i].Babylon.Constants) > 0 {
					for j := range store.Contracts[i].Babylon.Constants {
						relation := contract.ScriptConstants{
							ScriptId:         store.Contracts[i].BabylonID,
							GlobalConstantId: store.Contracts[i].Babylon.Constants[j].ID,
						}
						if _, err := tx.NewInsert().Model(&relation).Exec(ctx); err != nil {
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
				if err := store.Contracts[i].Jakarta.Save(ctx, tx); err != nil {
					return err
				}
				store.Contracts[i].JakartaID = store.Contracts[i].Jakarta.ID

				if len(store.Contracts[i].Jakarta.Constants) > 0 {
					for j := range store.Contracts[i].Jakarta.Constants {
						relation := contract.ScriptConstants{
							ScriptId:         store.Contracts[i].JakartaID,
							GlobalConstantId: store.Contracts[i].Jakarta.Constants[j].ID,
						}
						if _, err := tx.NewInsert().Model(&relation).Exec(ctx); err != nil {
							return err
						}
					}
				}

			} else {
				store.Contracts[i].JakartaID = store.Contracts[i].Babylon.ID
			}
		}

		if err := store.Contracts[i].Account.Save(ctx, tx); err != nil {
			return err
		}
		store.Contracts[i].AccountID = store.Contracts[i].Account.ID

		if !store.Contracts[i].Manager.IsEmpty() {
			if err := store.Contracts[i].Manager.Save(ctx, tx); err != nil {
				return err
			}
			store.Contracts[i].ManagerID = store.Contracts[i].Manager.ID
		}

		if !store.Contracts[i].Delegate.IsEmpty() {
			if err := store.Contracts[i].Delegate.Save(ctx, tx); err != nil {
				return err
			}
			store.Contracts[i].DelegateID = store.Contracts[i].Delegate.ID
		}
	}

	if _, err := tx.NewInsert().Model(&store.Contracts).Returning("id").Exec(ctx); err != nil {
		return err
	}

	return store.updateContracts(ctx, tx)
}

type contractUpdates struct {
	bun.BaseModel `bun:"contracts"`

	AccountID  int64
	LastAction time.Time
	TxCount    uint64
}

func (store *Store) updateContracts(ctx context.Context, tx bun.IDB) error {
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

	contracts := make([]*contractUpdates, 0, len(count))
	for accountID, txCount := range count {
		contracts = append(contracts, &contractUpdates{
			LastAction: store.Operations[0].Timestamp,
			AccountID:  accountID,
			TxCount:    txCount,
		})
	}

	values := tx.NewValues(&contracts)
	_, err := tx.NewUpdate().
		With("_data", values).
		Model((*contract.Contract)(nil)).
		TableExpr("_data").
		Set("last_action = _data.last_action").
		Set("tx_count = contract.tx_count + _data.tx_count").
		Where("contract.account_id = _data.account_id").
		Exec(ctx)
	return err
}
