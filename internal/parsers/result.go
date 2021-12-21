package parsers

import (
	"time"

	"github.com/baking-bad/bcdhub/internal/bcd"
	"github.com/baking-bad/bcdhub/internal/models/bigmapdiff"
	"github.com/baking-bad/bcdhub/internal/models/contract"
	"github.com/baking-bad/bcdhub/internal/models/global_constant"
	"github.com/baking-bad/bcdhub/internal/models/migration"
	"github.com/baking-bad/bcdhub/internal/models/operation"
	"github.com/baking-bad/bcdhub/internal/models/tokenbalance"
	"github.com/baking-bad/bcdhub/internal/models/types"
	"github.com/go-pg/pg/v10"
)

// Result -
type Result struct {
	// BigMapActions   []*bigmapaction.BigMapAction
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
		// BigMapActions:   make([]*bigmapaction.BigMapAction, 0),
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
	if len(result.Operations) > 0 {
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
			if len(result.Operations[i].Transfers) > 0 {
				if _, err := tx.Model(&result.Operations[i].Transfers).Returning("id").Insert(); err != nil {
					return err
				}
			}
			if len(result.Operations[i].BigMapActions) > 0 {
				if _, err := tx.Model(&result.Operations[i].BigMapActions).Returning("id").Insert(); err != nil {
					return err
				}
			}
		}
	}
	for i := range result.BigMapState {
		if err := result.BigMapState[i].Save(tx); err != nil {
			return err
		}
	}
	if len(result.Contracts) > 0 {
		if _, err := tx.Model(&result.Contracts).Returning("id").Insert(); err != nil {
			return err
		}
	}
	for i := range result.TokenBalances {
		if err := result.TokenBalances[i].Save(tx); err != nil {
			return err
		}
	}
	if len(result.Migrations) > 0 {
		if _, err := tx.Model(&result.Migrations).Returning("id").Insert(); err != nil {
			return err
		}
	}
	if len(result.GlobalConstants) > 0 {
		if _, err := tx.Model(&result.GlobalConstants).Returning("id").Insert(); err != nil {
			return err
		}
	}

	return result.updateContracts(tx)
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

type contractUpdates struct {
	//nolint
	tableName struct{} `pg:"contracts"`

	Network    types.Network
	Address    string
	LastAction time.Time
	TxCount    uint64
}

func (result *Result) updateContracts(tx pg.DBI) error {
	if len(result.Operations) == 0 {
		return nil
	}
	count := make(map[string]uint64)
	for i := range result.Operations {
		address := result.Operations[i].Destination
		if !bcd.IsContract(address) {
			continue
		}

		_, ok := count[address]
		if ok {
			count[address] += 1
		} else {
			count[address] = 1
		}
	}

	if len(count) == 0 {
		return nil
	}

	contracts := make([]*contractUpdates, 0, len(count))
	for address, txCount := range count {
		contracts = append(contracts, &contractUpdates{
			Network:    result.Operations[0].Network,
			LastAction: result.Operations[0].Timestamp,
			Address:    address,
			TxCount:    txCount,
		})
	}

	if _, err := tx.Model(&contracts).
		Set("last_action = _data.last_action, tx_count = contract_updates.tx_count + _data.tx_count").
		Where("contract_updates.address = _data.address").
		Where("contract_updates.network = _data.network").
		Update(); err != nil {
		return err
	}

	return nil
}
