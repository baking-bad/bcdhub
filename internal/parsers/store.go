package parsers

import (
	"context"
	"fmt"

	"github.com/baking-bad/bcdhub/internal/models/account"
	"github.com/baking-bad/bcdhub/internal/models/bigmapdiff"
	"github.com/baking-bad/bcdhub/internal/models/block"
	"github.com/baking-bad/bcdhub/internal/models/contract"
	"github.com/baking-bad/bcdhub/internal/models/migration"
	"github.com/baking-bad/bcdhub/internal/models/operation"
	smartrollup "github.com/baking-bad/bcdhub/internal/models/smart_rollup"
	"github.com/baking-bad/bcdhub/internal/models/ticket"
)

// Store -
type Store interface {
	AddBigMapStates(states ...*bigmapdiff.BigMapState)
	AddContracts(contracts ...*contract.Contract)
	AddMigrations(migrations ...*migration.Migration)
	AddOperations(operations ...*operation.Operation)
	AddGlobalConstants(constants ...*contract.GlobalConstant)
	AddSmartRollups(rollups ...*smartrollup.SmartRollup)
	AddTickets(tickets ...ticket.Ticket)
	AddTicketBalances(balances ...ticket.Balance)
	ListContracts() []*contract.Contract
	ListOperations() []*operation.Operation
	AddAccounts(accounts ...account.Account)
	Save(ctx context.Context) error
	SetBlock(block *block.Block)
}

// TestStore -
type TestStore struct {
	Block           *block.Block
	BigMapState     []*bigmapdiff.BigMapState
	Contracts       []*contract.Contract
	Migrations      []*migration.Migration
	Operations      []*operation.Operation
	GlobalConstants []*contract.GlobalConstant
	SmartRollups    []*smartrollup.SmartRollup
	Tickets         map[string]*ticket.Ticket
	TicketBalances  map[string]*ticket.Balance
	Accounts        map[string]*account.Account
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
		Tickets:         make(map[string]*ticket.Ticket, 0),
		TicketBalances:  make(map[string]*ticket.Balance, 0),
		Accounts:        make(map[string]*account.Account),
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

// AddAccounts -
func (store *TestStore) AddAccounts(accounts ...account.Account) {
	for i := range accounts {
		if account, ok := store.Accounts[accounts[i].Address]; !ok {
			store.Accounts[accounts[i].Address] = &accounts[i]
		} else {
			account.OperationsCount += accounts[i].OperationsCount
			account.EventsCount += accounts[i].EventsCount
			account.MigrationsCount += accounts[i].MigrationsCount
			account.TicketUpdatesCount += accounts[i].TicketUpdatesCount
		}
	}
}

// AddTickets -
func (store *TestStore) AddTickets(tickets ...ticket.Ticket) {
	for i := range tickets {
		hash := tickets[i].GetHash()
		if t, ok := store.Tickets[hash]; !ok {
			store.Tickets[hash] = &tickets[i]
		} else {
			t.UpdatesCount += tickets[i].UpdatesCount
		}
	}
}

// AddTicketBalances -
func (store *TestStore) AddTicketBalances(balance ...ticket.Balance) {
	for i := range balance {
		key := fmt.Sprintf("%s_%s", balance[i].Ticket.GetHash(), balance[i].Account.Address)
		if t, ok := store.TicketBalances[key]; !ok {
			store.TicketBalances[key] = &balance[i]
		} else {
			t.Amount = t.Amount.Add(balance[i].Amount)
		}
	}
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

func (store *TestStore) SetBlock(block *block.Block) {
	store.Block = block
}
