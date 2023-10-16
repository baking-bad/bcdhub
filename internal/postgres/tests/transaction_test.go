package tests

import (
	"context"
	"time"

	"github.com/baking-bad/bcdhub/internal/models/account"
	"github.com/baking-bad/bcdhub/internal/models/bigmapdiff"
	"github.com/baking-bad/bcdhub/internal/models/contract"
	"github.com/baking-bad/bcdhub/internal/models/migration"
	"github.com/baking-bad/bcdhub/internal/models/protocol"
	"github.com/baking-bad/bcdhub/internal/models/stats"
	"github.com/baking-bad/bcdhub/internal/models/ticket"
	"github.com/baking-bad/bcdhub/internal/models/types"
	"github.com/baking-bad/bcdhub/internal/postgres/core"
	"github.com/baking-bad/bcdhub/internal/testsuite"
	"github.com/shopspring/decimal"
)

func (s *StorageTestSuite) TestSave() {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	tx, err := core.NewTransaction(ctx, s.storage.DB)
	s.Require().NoError(err)

	account := account.Account{
		Address: "address",
		Type:    types.AccountTypeContract,
		Level:   100,
	}
	err = tx.Save(ctx, &account)
	s.Require().NoError(err)
	s.Require().Positive(account.ID)

	err = tx.Commit()
	s.Require().NoError(err)
}

func (s *StorageTestSuite) TestMigrations() {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	tx, err := core.NewTransaction(ctx, s.storage.DB)
	s.Require().NoError(err)

	m := migration.Migration{
		ProtocolID:     1,
		PrevProtocolID: 0,
		Hash:           []byte{0, 1, 2, 3, 4},
		Timestamp:      time.Now(),
		Level:          100,
		Kind:           types.MigrationKindBootstrap,
		ContractID:     1,
	}
	err = tx.Migrations(ctx, &m)
	s.Require().NoError(err)
	s.Require().Positive(m.ID)

	err = tx.Commit()
	s.Require().NoError(err)
}

func (s *StorageTestSuite) TestProtocol() {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	tx, err := core.NewTransaction(ctx, s.storage.DB)
	s.Require().NoError(err)

	p := protocol.Protocol{
		Hash:       "protocol_hash",
		StartLevel: 100,
		EndLevel:   200,
		SymLink:    "symlink",
		Alias:      "alias",
		ChainID:    "chain_id",
		Constants:  &protocol.Constants{},
	}
	err = tx.Protocol(ctx, &p)
	s.Require().NoError(err)
	s.Require().Positive(p.ID)

	err = tx.Commit()
	s.Require().NoError(err)
}

func (s *StorageTestSuite) TestScriptConstants() {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	tx, err := core.NewTransaction(ctx, s.storage.DB)
	s.Require().NoError(err)

	sc := []*contract.ScriptConstants{
		{
			ScriptId:         1,
			GlobalConstantId: 1,
		}, {
			ScriptId:         2,
			GlobalConstantId: 1,
		}, {
			ScriptId:         1,
			GlobalConstantId: 2,
		},
	}
	err = tx.ScriptConstant(ctx, sc...)
	s.Require().NoError(err)

	err = tx.Commit()
	s.Require().NoError(err)
}

func (s *StorageTestSuite) TestScripts() {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	tx, err := core.NewTransaction(ctx, s.storage.DB)
	s.Require().NoError(err)

	sc := []*contract.Script{
		{
			Hash: "hash_1",
		}, {
			Hash: "hash_2",
		}, {
			Hash: "hash_3",
		},
	}
	err = tx.Scripts(ctx, sc...)
	s.Require().NoError(err)

	err = tx.Commit()
	s.Require().NoError(err)
}

func (s *StorageTestSuite) TestScriptsConflict() {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	tx, err := core.NewTransaction(ctx, s.storage.DB)
	s.Require().NoError(err)

	update := contract.Script{
		Hash: "8436dde35bd56644cd4f40c5f26839cb8f4b51052e415da2b9fadcd9bddcb03e",
	}
	err = tx.Scripts(ctx, &update)
	s.Require().NoError(err)
	s.Require().EqualValues(5, update.ID)

	err = tx.Commit()
	s.Require().NoError(err)
}

func (s *StorageTestSuite) TestAccounts() {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	tx, err := core.NewTransaction(ctx, s.storage.DB)
	s.Require().NoError(err)

	sc := []*account.Account{
		{
			Address:            "address_1",
			Type:               types.AccountTypeContract,
			Level:              100,
			TicketUpdatesCount: 2,
		}, {
			Address:         "address_12",
			Type:            types.AccountTypeSmartRollup,
			Level:           100,
			MigrationsCount: 2,
		}, {
			Address:     "address_2",
			Type:        types.AccountTypeTz,
			Level:       100,
			EventsCount: 2,
		},
	}
	err = tx.Accounts(ctx, sc...)
	s.Require().NoError(err)

	err = tx.Commit()
	s.Require().NoError(err)
}

func (s *StorageTestSuite) TestBigMapStates() {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	tx, err := core.NewTransaction(ctx, s.storage.DB)
	s.Require().NoError(err)

	sc := []*bigmapdiff.BigMapState{
		{
			Key:             []byte{0, 1, 2, 3},
			KeyHash:         "hash 1",
			Ptr:             100000,
			LastUpdateLevel: 100,
			Count:           1,
			Removed:         false,
			Contract:        "contract 1",
		}, {
			Key:             []byte{0, 1, 2, 3, 4},
			KeyHash:         "hash 2",
			Ptr:             100000,
			LastUpdateLevel: 100,
			Count:           1,
			Removed:         false,
			Contract:        "contract 2",
		}, {
			Key:             []byte{0, 1, 2, 3, 5},
			KeyHash:         "hash 3",
			Ptr:             100000,
			LastUpdateLevel: 100,
			Count:           1,
			Removed:         false,
			Contract:        "contract 3"},
	}
	err = tx.BigMapStates(ctx, sc...)
	s.Require().NoError(err)

	err = tx.Commit()
	s.Require().NoError(err)

	var result []bigmapdiff.BigMapState
	err = s.storage.DB.NewSelect().Model(&result).Where("ptr = 100000").Scan(ctx)
	s.Require().NoError(err)
	s.Require().Len(result, 3)
}

func (s *StorageTestSuite) TestBabylonUpdateNonDelegator() {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	tx, err := core.NewTransaction(ctx, s.storage.DB)
	s.Require().NoError(err)

	c := contract.Contract{
		ID:        2,
		BabylonID: 10,
	}

	err = tx.BabylonUpdateNonDelegator(ctx, &c)
	s.Require().NoError(err)

	err = tx.Commit()
	s.Require().NoError(err)

	var newContract contract.Contract
	err = s.storage.DB.NewSelect().Model(&newContract).Where("id = 2").Scan(ctx)
	s.Require().NoError(err)
	s.Require().EqualValues(10, newContract.BabylonID)
}

func (s *StorageTestSuite) TestJakartaVesting() {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	tx, err := core.NewTransaction(ctx, s.storage.DB)
	s.Require().NoError(err)

	c := contract.Contract{
		ID: 2,
	}

	err = tx.JakartaVesting(ctx, &c)
	s.Require().NoError(err)

	err = tx.Commit()
	s.Require().NoError(err)

	var newContract contract.Contract
	err = s.storage.DB.NewSelect().Model(&newContract).Where("id = 2").Scan(ctx)
	s.Require().NoError(err)
	s.Require().EqualValues(5, newContract.JakartaID)
}

func (s *StorageTestSuite) TestJakartaUpdateNonDelegator() {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	tx, err := core.NewTransaction(ctx, s.storage.DB)
	s.Require().NoError(err)

	c := contract.Contract{
		ID:        2,
		JakartaID: 100,
	}

	err = tx.JakartaUpdateNonDelegator(ctx, &c)
	s.Require().NoError(err)

	err = tx.Commit()
	s.Require().NoError(err)

	var newContract contract.Contract
	err = s.storage.DB.NewSelect().Model(&newContract).Where("id = 2").Scan(ctx)
	s.Require().NoError(err)
	s.Require().EqualValues(100, newContract.JakartaID)
}

func (s *StorageTestSuite) TestToJakarta() {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	tx, err := core.NewTransaction(ctx, s.storage.DB)
	s.Require().NoError(err)

	err = tx.ToJakarta(ctx)
	s.Require().NoError(err)

	err = tx.Commit()
	s.Require().NoError(err)

	var newContract contract.Contract
	err = s.storage.DB.NewSelect().Model(&newContract).Where("id = 16").Scan(ctx)
	s.Require().NoError(err)
	s.Require().EqualValues(14, newContract.JakartaID)
}

func (s *StorageTestSuite) TestDeleteBigMapStates() {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	tx, err := core.NewTransaction(ctx, s.storage.DB)
	s.Require().NoError(err)

	states, err := tx.DeleteBigMapStatesByContract(ctx, "KT1Pz65ssbPF7Zv9Dh7ggqUkgAYNSuJ9iia7")
	s.Require().NoError(err)

	err = tx.Commit()
	s.Require().NoError(err)
	s.Require().Len(states, 9)

	s.Require().Equal("KT1Pz65ssbPF7Zv9Dh7ggqUkgAYNSuJ9iia7", states[0].Contract)
}

func (s *StorageTestSuite) TestUpdateStats() {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	tx, err := core.NewTransaction(ctx, s.storage.DB)
	s.Require().NoError(err)

	err = tx.UpdateStats(ctx, stats.Stats{
		ID:                  1,
		ContractsCount:      1,
		OperationsCount:     4,
		OriginationsCount:   1,
		TransactionsCount:   1,
		EventsCount:         1,
		SrOriginationsCount: 1,
	})
	s.Require().NoError(err)

	err = tx.Commit()
	s.Require().NoError(err)

	var stats stats.Stats
	err = s.storage.DB.NewSelect().Model(&stats).Limit(1).Scan(ctx)
	s.Require().NoError(err)
	s.Require().EqualValues(121, stats.ContractsCount)
	s.Require().EqualValues(196, stats.OperationsCount)
	s.Require().EqualValues(73, stats.TransactionsCount)
	s.Require().EqualValues(119, stats.OriginationsCount)
	s.Require().EqualValues(3, stats.EventsCount)
	s.Require().EqualValues(1, stats.SrOriginationsCount)
}

func (s *StorageTestSuite) TestTickets() {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	tx, err := core.NewTransaction(ctx, s.storage.DB)
	s.Require().NoError(err)

	err = tx.Tickets(ctx, &ticket.Ticket{
		ContentType:  testsuite.MustHexDecode("7b227072696d223a22737472696e67227d"),
		Content:      testsuite.MustHexDecode("7b22737472696e67223a22616263227d"),
		TicketerID:   133,
		UpdatesCount: 1,
	}, &ticket.Ticket{
		ContentType:  testsuite.MustHexDecode("7b227072696d223a22737472696e67227d"),
		Content:      testsuite.MustHexDecode("7b22737472696e67223a22616263227d"),
		TicketerID:   132,
		UpdatesCount: 2,
	})
	s.Require().NoError(err)

	err = tx.Commit()
	s.Require().NoError(err)

	var tickets []ticket.Ticket
	err = s.storage.DB.NewSelect().Model(&tickets).Scan(ctx)
	s.Require().NoError(err)
	s.Require().Len(tickets, 3)
}

func (s *StorageTestSuite) TestTicketBalances() {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	tx, err := core.NewTransaction(ctx, s.storage.DB)
	s.Require().NoError(err)

	err = tx.TicketBalances(ctx, &ticket.Balance{
		TicketId:  1,
		AccountId: 131,
		Amount:    decimal.RequireFromString("17"),
	}, &ticket.Balance{
		TicketId:  1,
		AccountId: 10,
		Amount:    decimal.RequireFromString("1"),
	})
	s.Require().NoError(err)

	err = tx.Commit()
	s.Require().NoError(err)

	var balances []ticket.Balance
	err = s.storage.DB.NewSelect().Model(&balances).Where("ticket_id = 1").Scan(ctx)
	s.Require().NoError(err)
	s.Require().Len(balances, 3)
}

func (s *StorageTestSuite) TestBabylonUpdateBigMapDiffs() {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	tx, err := core.NewTransaction(ctx, s.storage.DB)
	s.Require().NoError(err)

	count, err := tx.BabylonUpdateBigMapDiffs(ctx, "KT1R9BdHMfGTwKnbCHii8akcB7DqzfdnD9AD", 10000)
	s.Require().NoError(err)

	err = tx.Commit()
	s.Require().NoError(err)

	s.Require().EqualValues(4, count)

	var diffs []bigmapdiff.BigMapDiff
	err = s.storage.DB.NewSelect().Model(&diffs).Where("ptr = 10000").Scan(ctx)
	s.Require().NoError(err)
	s.Require().Len(diffs, 4)
}
