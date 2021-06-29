package parsers

import (
	"github.com/baking-bad/bcdhub/internal/bcd"
	"github.com/baking-bad/bcdhub/internal/models/bigmap"
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
	BigMaps       []*bigmap.BigMap
	BigMapState   []*bigmap.State
	Contracts     []*contract.Contract
	Migrations    []*migration.Migration
	Operations    []*operation.Operation
	TokenBalances []*tokenbalance.TokenBalance
}

// NewResult -
func NewResult() *Result {
	return &Result{
		BigMaps:       make([]*bigmap.BigMap, 0),
		BigMapState:   make([]*bigmap.State, 0),
		Contracts:     make([]*contract.Contract, 0),
		Migrations:    make([]*migration.Migration, 0),
		Operations:    make([]*operation.Operation, 0),
		TokenBalances: make([]*tokenbalance.TokenBalance, 0),
	}
}

// Save -
func (result *Result) Save(tx *gorm.DB) error {
	if err := tx.CreateInBatches(result.BigMaps, batchSize).Error; err != nil {
		return err
	}

	result.fillBigMapIDs()

	if err := tx.CreateInBatches(result.Operations, batchSize).Error; err != nil {
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

// Clear -
func (result *Result) Clear() {
	result.BigMaps = nil
	result.BigMapState = nil
	result.Contracts = nil
	result.Migrations = nil
	result.Operations = nil
	result.TokenBalances = nil
}

// Merge -
func (result *Result) Merge(second *Result) {
	if second == nil {
		return
	}

	result.BigMaps = append(result.BigMaps, second.BigMaps...)
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

type key struct {
	Ptr     int64
	Address string
	Network int64
}

func (result *Result) fillBigMapIDs() {
	if len(result.BigMaps) == 0 {
		return
	}

	if len(result.Operations) == 0 && len(result.BigMapState) == 0 {
		return
	}

	maps := make(map[key]int64)
	for i := range result.BigMaps {
		maps[key{
			Ptr:     result.BigMaps[i].Ptr,
			Address: result.BigMaps[i].Contract,
			Network: int64(result.BigMaps[i].Network),
		}] = result.BigMaps[i].ID
	}

	for i := range result.BigMapState {
		if result.BigMapState[i].BigMapID > 0 {
			continue
		}
		if id, ok := maps[key{
			Ptr:     result.BigMapState[i].BigMap.Ptr,
			Address: result.BigMapState[i].BigMap.Contract,
			Network: int64(result.BigMapState[i].BigMap.Network),
		}]; ok {
			result.BigMapState[i].BigMapID = id
		}
	}

	for i := range result.Operations {
		if len(result.Operations[i].BigMapActions) == 0 && len(result.Operations[i].BigMapDiffs) == 0 {
			continue
		}
		for j, diff := range result.Operations[i].BigMapDiffs {
			if diff.BigMapID > 0 {
				continue
			}
			if id, ok := maps[key{
				Ptr:     diff.BigMap.Ptr,
				Address: diff.BigMap.Contract,
				Network: int64(diff.BigMap.Network),
			}]; ok {
				result.Operations[i].BigMapDiffs[j].BigMapID = id
			}
		}

		for j, action := range result.Operations[i].BigMapActions {
			if action.DestinationID != nil && *action.DestinationID > -1 {
				if id, ok := maps[key{
					Ptr:     action.Destination.Ptr,
					Address: action.Destination.Contract,
					Network: int64(action.Destination.Network),
				}]; ok {
					result.Operations[i].BigMapActions[j].DestinationID = &id
				}
			}
			if action.SourceID != nil && *action.SourceID > -1 {
				if id, ok := maps[key{
					Ptr:     action.Source.Ptr,
					Address: action.Source.Contract,
					Network: int64(action.Source.Network),
				}]; ok {
					result.Operations[i].BigMapActions[j].SourceID = &id
				}
			}
		}
	}
}
