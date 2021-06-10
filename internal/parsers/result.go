package parsers

import (
	"github.com/baking-bad/bcdhub/internal/bcd"
	"github.com/baking-bad/bcdhub/internal/models/bigmapaction"
	"github.com/baking-bad/bcdhub/internal/models/bigmapdiff"
	"github.com/baking-bad/bcdhub/internal/models/contract"
	"github.com/baking-bad/bcdhub/internal/models/migration"
	"github.com/baking-bad/bcdhub/internal/models/operation"
	"github.com/baking-bad/bcdhub/internal/models/tokenbalance"
	"gorm.io/gorm"
)

const (
	batchSize = 1000
)

// Result -
type Result struct {
	BigMapActions []*bigmapaction.BigMapAction
	BigMapState   []*bigmapdiff.BigMapState
	Contracts     []*contract.Contract
	Migrations    []*migration.Migration
	Operations    []*operation.Operation
	TokenBalances []*tokenbalance.TokenBalance
}

// NewResult -
func NewResult() *Result {
	return &Result{
		BigMapActions: make([]*bigmapaction.BigMapAction, 0),
		BigMapState:   make([]*bigmapdiff.BigMapState, 0),
		Contracts:     make([]*contract.Contract, 0),
		Migrations:    make([]*migration.Migration, 0),
		Operations:    make([]*operation.Operation, 0),
		TokenBalances: make([]*tokenbalance.TokenBalance, 0),
	}
}

// Save -
func (result *Result) Save(tx *gorm.DB) error {
	if err := tx.CreateInBatches(result.Operations, batchSize).Error; err != nil {
		return err
	}
	if err := tx.CreateInBatches(result.BigMapActions, batchSize).Error; err != nil {
		return err
	}
	for i := range result.BigMapState {
		if err := result.BigMapState[i].Save(tx); err != nil {
			return err
		}
	}
	if err := tx.CreateInBatches(result.Contracts, batchSize).Error; err != nil {
		return err
	}
	for i := range result.TokenBalances {
		if err := result.TokenBalances[i].Save(tx); err != nil {
			return err
		}
	}
	if err := tx.CreateInBatches(result.Migrations, batchSize).Error; err != nil {
		return err
	}

	return result.updateContracts(tx)
}

// Merge -
func (result *Result) Merge(second *Result) {
	if second == nil {
		return
	}

	result.BigMapActions = append(result.BigMapActions, second.BigMapActions...)
	result.BigMapState = append(result.BigMapState, second.BigMapState...)
	result.Contracts = append(result.Contracts, second.Contracts...)
	result.Migrations = append(result.Migrations, second.Migrations...)
	result.Operations = append(result.Operations, second.Operations...)
	result.TokenBalances = append(result.TokenBalances, second.TokenBalances...)
}

func (result *Result) updateContracts(tx *gorm.DB) error {
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

	network := result.Operations[0].Network
	ts := result.Operations[0].Timestamp

	for address, txCount := range count {
		if err := tx.Exec(`UPDATE contracts SET last_action = ?, tx_count = tx_count + ? WHERE address = ? AND network = ?;`, ts, txCount, address, network).Error; err != nil {
			return err
		}
	}

	return nil
}
