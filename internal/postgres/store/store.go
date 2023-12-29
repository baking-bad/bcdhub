package store

import (
	"fmt"

	"github.com/baking-bad/bcdhub/internal/models/account"
	"github.com/baking-bad/bcdhub/internal/models/bigmapdiff"
	"github.com/baking-bad/bcdhub/internal/models/block"
	"github.com/baking-bad/bcdhub/internal/models/contract"
	"github.com/baking-bad/bcdhub/internal/models/migration"
	"github.com/baking-bad/bcdhub/internal/models/operation"
	smartrollup "github.com/baking-bad/bcdhub/internal/models/smart_rollup"
	"github.com/baking-bad/bcdhub/internal/models/stats"
	"github.com/baking-bad/bcdhub/internal/models/ticket"
	"github.com/uptrace/bun"
)

// Store -
type Store struct {
	Block           *block.Block
	BigMapState     map[string]*bigmapdiff.BigMapState
	Contracts       []*contract.Contract
	Migrations      []*migration.Migration
	Operations      []*operation.Operation
	GlobalConstants []*contract.GlobalConstant
	SmartRollups    []*smartrollup.SmartRollup
	Tickets         map[string]*ticket.Ticket
	TicketBalances  map[string]*ticket.Balance
	Accounts        map[string]*account.Account
	Stats           stats.Stats

	stats     stats.Repository
	db        *bun.DB
	accIds    map[string]int64
	ticketIds map[string]int64
}

// NewStore -
func NewStore(db *bun.DB, statsRepo stats.Repository) *Store {
	return &Store{
		BigMapState:     make(map[string]*bigmapdiff.BigMapState),
		Contracts:       make([]*contract.Contract, 0),
		Migrations:      make([]*migration.Migration, 0),
		Operations:      make([]*operation.Operation, 0),
		GlobalConstants: make([]*contract.GlobalConstant, 0),
		SmartRollups:    make([]*smartrollup.SmartRollup, 0),
		Tickets:         make(map[string]*ticket.Ticket),
		TicketBalances:  make(map[string]*ticket.Balance),
		Accounts:        make(map[string]*account.Account),
		Stats:           stats.Stats{},
		stats:           statsRepo,
		db:              db,
		accIds:          make(map[string]int64),
		ticketIds:       make(map[string]int64),
	}
}

func (store *Store) SetBlock(block *block.Block) {
	store.Block = block
}

// AddBigMapStates -
func (store *Store) AddBigMapStates(states ...*bigmapdiff.BigMapState) {
	for i := range states {
		key := states[i].String()
		if state, ok := store.BigMapState[key]; ok {
			state.Count += 1
			state.Value = states[i].Value
			state.Removed = states[i].Removed
		} else {
			store.BigMapState[key] = states[i]
		}
	}
}

func (store *Store) bigMapStates() []*bigmapdiff.BigMapState {
	arr := make([]*bigmapdiff.BigMapState, 0, len(store.BigMapState))
	for _, state := range store.BigMapState {
		arr = append(arr, state)
	}
	return arr
}

// AddContracts -
func (store *Store) AddContracts(contracts ...*contract.Contract) {
	store.Contracts = append(store.Contracts, contracts...)

	store.Stats.ContractsCount += len(contracts)
}

// AddMigrations -
func (store *Store) AddMigrations(migrations ...*migration.Migration) {
	store.Migrations = append(store.Migrations, migrations...)

	for i := range migrations {
		if migrations[i].Contract.Account.IsEmpty() {
			continue
		}

		if account, ok := store.Accounts[migrations[i].Contract.Account.Address]; !ok {
			store.Accounts[migrations[i].Contract.Account.Address] = &migrations[i].Contract.Account
		} else {
			account.MigrationsCount += 1
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
	store.Stats.SmartRollupsCount += len(rollups)
}

// AddAccounts -
func (store *Store) AddAccounts(accounts ...account.Account) {
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
func (store *Store) AddTickets(tickets ...ticket.Ticket) {
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
func (store *Store) AddTicketBalances(balance ...ticket.Balance) {
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
func (store *Store) ListContracts() []*contract.Contract {
	return store.Contracts
}

// ListOperations -
func (store *Store) ListOperations() []*operation.Operation {
	return store.Operations
}
