package parsers

import (
	"context"

	"github.com/baking-bad/bcdhub/internal/models/bigmapdiff"
	"github.com/baking-bad/bcdhub/internal/models/contract"
	"github.com/baking-bad/bcdhub/internal/models/migration"
	"github.com/baking-bad/bcdhub/internal/models/operation"
	smartrollup "github.com/baking-bad/bcdhub/internal/models/smart_rollup"
)

// Store -
type Store interface {
	AddBigMapStates(states ...*bigmapdiff.BigMapState)
	AddContracts(contracts ...*contract.Contract)
	AddMigrations(migrations ...*migration.Migration)
	AddOperations(operations ...*operation.Operation)
	AddGlobalConstants(constants ...*contract.GlobalConstant)
	AddSmartRollups(rollups ...*smartrollup.SmartRollup)
	ListContracts() []*contract.Contract
	ListOperations() []*operation.Operation
	Save(ctx context.Context) error
}

// TestStore -
type TestStore struct {
	BigMapState     []*bigmapdiff.BigMapState
	Contracts       []*contract.Contract
	Migrations      []*migration.Migration
	Operations      []*operation.Operation
	GlobalConstants []*contract.GlobalConstant
	SmartRollups    []*smartrollup.SmartRollup
}

// NewTestStore -
func NewTestStore() *TestStore {
	return &TestStore{
		BigMapState:     make([]*bigmapdiff.BigMapState, 0),
		Contracts:       make([]*contract.Contract, 0),
		Migrations:      make([]*migration.Migration, 0),
		Operations:      make([]*operation.Operation, 0),
		GlobalConstants: make([]*contract.GlobalConstant, 0),
		SmartRollups:    make([]*smartrollup.SmartRollup, 0),
	}
}

// AddBigMapStates -
func (store *TestStore) AddBigMapStates(states ...*bigmapdiff.BigMapState) {
	store.BigMapState = append(store.BigMapState, states...)
}

// AddContracts -
func (store *TestStore) AddContracts(contracts ...*contract.Contract) {
	store.Contracts = append(store.Contracts, contracts...)
}

// AddMigrations -
func (store *TestStore) AddMigrations(migrations ...*migration.Migration) {
	store.Migrations = append(store.Migrations, migrations...)
}

// AddOperations -
func (store *TestStore) AddOperations(operations ...*operation.Operation) {
	store.Operations = append(store.Operations, operations...)
}

// AddGlobalConstants -
func (store *TestStore) AddGlobalConstants(constants ...*contract.GlobalConstant) {
	store.GlobalConstants = append(store.GlobalConstants, constants...)
}

// AddSmartRollups -
func (store *TestStore) AddSmartRollups(rollups ...*smartrollup.SmartRollup) {
	store.SmartRollups = append(store.SmartRollups, rollups...)
}

// ListContracts -
func (store *TestStore) ListContracts() []*contract.Contract {
	return store.Contracts
}

// ListOperations -
func (store *TestStore) ListOperations() []*operation.Operation {
	return store.Operations
}

// Save -
func (store *TestStore) Save(ctx context.Context) error {
	return nil
}
