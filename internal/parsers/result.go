package parsers

import (
	"time"

	"github.com/baking-bad/bcdhub/internal/models/bigmapdiff"
	"github.com/baking-bad/bcdhub/internal/models/contract"
	"github.com/baking-bad/bcdhub/internal/models/global_constant"
	"github.com/baking-bad/bcdhub/internal/models/migration"
	"github.com/baking-bad/bcdhub/internal/models/operation"
	"github.com/baking-bad/bcdhub/internal/models/tokenbalance"
	"github.com/baking-bad/bcdhub/internal/models/transfer"
	"github.com/baking-bad/bcdhub/internal/models/types"
	"github.com/go-pg/pg/v10"
)

// Result -
type Result struct {
	BigMapState     []*bigmapdiff.BigMapState
	Contracts       []*contract.Contract
	Migrations      []*migration.Migration
	Operations      []*operation.Operation
	TokenBalances   []*tokenbalance.TokenBalance
	GlobalConstants []*global_constant.GlobalConstant
}

// NewResult -
func NewResult() *Result {
	return &Result{
		BigMapState:     make([]*bigmapdiff.BigMapState, 0),
		Contracts:       make([]*contract.Contract, 0),
		Migrations:      make([]*migration.Migration, 0),
		Operations:      make([]*operation.Operation, 0),
		TokenBalances:   make([]*tokenbalance.TokenBalance, 0),
		GlobalConstants: make([]*global_constant.GlobalConstant, 0),
	}
}

// Save -
func (result *Result) Save(tx pg.DBI) error {
	if err := result.saveOperations(tx); err != nil {
		return err
	}

	if err := result.saveContracts(tx); err != nil {
		return err
	}

	if err := result.saveMigrations(tx); err != nil {
		return err
	}

	for i := range result.BigMapState {
		if err := result.BigMapState[i].Save(tx); err != nil {
			return err
		}
	}

	if err := result.saveTokenBalances(tx); err != nil {
		return err
	}

	if len(result.GlobalConstants) > 0 {
		if _, err := tx.Model(&result.GlobalConstants).Returning("id").Insert(); err != nil {
			return err
		}
	}

	return nil
}

// Merge -
func (result *Result) Merge(second *Result) {
	if second == nil {
		return
	}

	result.BigMapState = append(result.BigMapState, second.BigMapState...)
	result.Contracts = append(result.Contracts, second.Contracts...)
	result.Migrations = append(result.Migrations, second.Migrations...)
	result.Operations = append(result.Operations, second.Operations...)
	result.TokenBalances = append(result.TokenBalances, second.TokenBalances...)
	result.GlobalConstants = append(result.GlobalConstants, second.GlobalConstants...)
}

func (result *Result) saveMigrations(tx pg.DBI) error {
	if len(result.Migrations) == 0 {
		return nil
	}

	for i := range result.Migrations {
		if result.Migrations[i].ContractID == 0 {
			result.Migrations[i].ContractID = result.Migrations[i].Contract.ID
		}
	}

	_, err := tx.Model(&result.Migrations).Returning("id").Insert()
	return err
}

func (result *Result) saveOperations(tx pg.DBI) error {
	if len(result.Operations) == 0 {
		return nil
	}

	for i := range result.Operations {
		if !result.Operations[i].Source.IsEmpty() {
			if err := result.Operations[i].Source.Save(tx); err != nil {
				return err
			}
			result.Operations[i].SourceID = result.Operations[i].Source.ID
		}
		if !result.Operations[i].Destination.IsEmpty() {
			if err := result.Operations[i].Destination.Save(tx); err != nil {
				return err
			}
			result.Operations[i].DestinationID = result.Operations[i].Destination.ID
		}
		if !result.Operations[i].Initiator.IsEmpty() {
			if err := result.Operations[i].Initiator.Save(tx); err != nil {
				return err
			}
			result.Operations[i].InitiatorID = result.Operations[i].Initiator.ID
		}
		if !result.Operations[i].Delegate.IsEmpty() {
			if err := result.Operations[i].Delegate.Save(tx); err != nil {
				return err
			}
			result.Operations[i].DelegateID = result.Operations[i].Delegate.ID
		}
	}

	if _, err := tx.Model(&result.Operations).Returning("id").Insert(); err != nil {
		return err
	}

	for i := range result.Operations {
		for j := range result.Operations[i].BigMapDiffs {
			result.Operations[i].BigMapDiffs[j].OperationID = result.Operations[i].ID
		}
		for j := range result.Operations[i].Transfers {
			result.Operations[i].Transfers[j].OperationID = result.Operations[i].ID
		}
		for j := range result.Operations[i].BigMapActions {
			result.Operations[i].BigMapActions[j].OperationID = result.Operations[i].ID
		}

		if len(result.Operations[i].BigMapDiffs) > 0 {
			if _, err := tx.Model(&result.Operations[i].BigMapDiffs).Returning("id").Insert(); err != nil {
				return err
			}
		}

		if err := result.saveTransfers(tx, result.Operations[i].Transfers); err != nil {
			return err
		}

		if len(result.Operations[i].BigMapActions) > 0 {
			if _, err := tx.Model(&result.Operations[i].BigMapActions).Returning("id").Insert(); err != nil {
				return err
			}
		}
	}
	return result.updateContracts(tx)
}

func (result *Result) saveContracts(tx pg.DBI) error {
	if len(result.Contracts) == 0 {
		return nil
	}

	for i := range result.Contracts {
		if result.Contracts[i].Alpha.Code != nil {
			if err := result.Contracts[i].Alpha.Save(tx); err != nil {
				return err
			}
			result.Contracts[i].AlphaID = result.Contracts[i].Alpha.ID
		}
		if result.Contracts[i].Babylon.Code != nil {
			if result.Contracts[i].Alpha.Hash != result.Contracts[i].Babylon.Hash {
				if err := result.Contracts[i].Babylon.Save(tx); err != nil {
					return err
				}
				result.Contracts[i].BabylonID = result.Contracts[i].Babylon.ID
			} else {
				result.Contracts[i].BabylonID = result.Contracts[i].Alpha.ID
			}
		}

		if err := result.Contracts[i].Account.Save(tx); err != nil {
			return err
		}
		result.Contracts[i].AccountID = result.Contracts[i].Account.ID

		if !result.Contracts[i].Manager.IsEmpty() {
			if err := result.Contracts[i].Manager.Save(tx); err != nil {
				return err
			}
			result.Contracts[i].ManagerID = result.Contracts[i].Manager.ID
		}

		if !result.Contracts[i].Delegate.IsEmpty() {
			if err := result.Contracts[i].Delegate.Save(tx); err != nil {
				return err
			}
			result.Contracts[i].DelegateID = result.Contracts[i].Delegate.ID
		}
	}

	if _, err := tx.Model(&result.Contracts).Returning("id").Insert(); err != nil {
		return err
	}

	return result.updateContracts(tx)
}

type contractUpdates struct {
	//nolint
	tableName struct{} `pg:"contracts"`

	Network    types.Network
	AccountID  int64
	LastAction time.Time
	TxCount    uint64
}

func (result *Result) updateContracts(tx pg.DBI) error {
	if len(result.Operations) == 0 {
		return nil
	}
	count := make(map[int64]uint64)
	for i := range result.Operations {
		destination := result.Operations[i].Destination
		if destination.Type != types.AccountTypeContract {
			continue
		}

		if value, ok := count[destination.ID]; ok {
			count[destination.ID] = value + 1
		} else {
			count[destination.ID] = 1
		}

		source := result.Operations[i].Source
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
			Network:    result.Operations[0].Network,
			LastAction: result.Operations[0].Timestamp,
			AccountID:  accountID,
			TxCount:    txCount,
		})
	}

	_, err := tx.Model(&contracts).
		Set("last_action = ?last_action, tx_count = contract_updates.tx_count + ?tx_count").
		Where("contract_updates.account_id = ?account_id").
		Where("contract_updates.network = ?network").
		Update()
	return err
}

func (result *Result) saveTokenBalances(tx pg.DBI) error {
	for i := range result.TokenBalances {
		if result.TokenBalances[i].AccountID == 0 {
			if err := result.TokenBalances[i].Account.Save(tx); err != nil {
				return err
			}
			result.TokenBalances[i].AccountID = result.TokenBalances[i].Account.ID
		}
		if err := result.TokenBalances[i].Save(tx); err != nil {
			return err
		}
	}
	return nil
}

func (result *Result) saveTransfers(tx pg.DBI, transfers []*transfer.Transfer) error {
	if len(transfers) == 0 {
		return nil
	}
	for j := range transfers {
		if !transfers[j].Initiator.IsEmpty() {
			if err := transfers[j].Initiator.Save(tx); err != nil {
				return err
			}
			transfers[j].InitiatorID = transfers[j].Initiator.ID
		}
		if !transfers[j].From.IsEmpty() {
			if err := transfers[j].From.Save(tx); err != nil {
				return err
			}
			transfers[j].FromID = transfers[j].From.ID
		}
		if !transfers[j].To.IsEmpty() {
			if err := transfers[j].To.Save(tx); err != nil {
				return err
			}
			transfers[j].ToID = transfers[j].To.ID
		}
	}

	_, err := tx.Model(&transfers).Returning("id").Insert()
	return err
}
