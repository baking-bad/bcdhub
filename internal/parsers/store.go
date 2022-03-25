package parsers

import (
	"github.com/baking-bad/bcdhub/internal/models/bigmapdiff"
	"github.com/baking-bad/bcdhub/internal/models/contract"
	"github.com/baking-bad/bcdhub/internal/models/global_constant"
	"github.com/baking-bad/bcdhub/internal/models/migration"
	"github.com/baking-bad/bcdhub/internal/models/operation"
	"github.com/baking-bad/bcdhub/internal/models/tokenbalance"
)

// Store -
type Store interface {
	AddBigMapStates(states ...*bigmapdiff.BigMapState)
	AddContracts(contracts ...*contract.Contract)
	AddMigrations(migrations ...*migration.Migration)
	AddOperations(operations ...*operation.Operation)
	AddTokenBalances(balances ...*tokenbalance.TokenBalance)
	AddGlobalConstants(constants ...*global_constant.GlobalConstant)
	ListContracts() []*contract.Contract
	ListOperations() []*operation.Operation
	Save() error
}

// TestStore -
type TestStore struct {
	BigMapState     []*bigmapdiff.BigMapState
	Contracts       []*contract.Contract
	Migrations      []*migration.Migration
	Operations      []*operation.Operation
	TokenBalances   []*tokenbalance.TokenBalance
	GlobalConstants []*global_constant.GlobalConstant
}

// NewTestStore -
func NewTestStore() *TestStore {
	return &TestStore{
		BigMapState:     make([]*bigmapdiff.BigMapState, 0),
		Contracts:       make([]*contract.Contract, 0),
		Migrations:      make([]*migration.Migration, 0),
		Operations:      make([]*operation.Operation, 0),
		TokenBalances:   make([]*tokenbalance.TokenBalance, 0),
		GlobalConstants: make([]*global_constant.GlobalConstant, 0),
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

// AddTokenBalances -
func (store *TestStore) AddTokenBalances(balances ...*tokenbalance.TokenBalance) {
	store.TokenBalances = append(store.TokenBalances, balances...)
}

// AddGlobalConstants -
func (store *TestStore) AddGlobalConstants(constants ...*global_constant.GlobalConstant) {
	store.GlobalConstants = append(store.GlobalConstants, constants...)
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
func (store *TestStore) Save() error {
	return nil
}
