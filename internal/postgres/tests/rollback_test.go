package tests

import (
	"context"
	"time"

	"github.com/baking-bad/bcdhub/internal/models/account"
	"github.com/baking-bad/bcdhub/internal/models/bigmapdiff"
	"github.com/baking-bad/bcdhub/internal/models/block"
	"github.com/baking-bad/bcdhub/internal/models/protocol"
	"github.com/baking-bad/bcdhub/internal/models/stats"
	"github.com/baking-bad/bcdhub/internal/models/ticket"
	"github.com/baking-bad/bcdhub/internal/models/types"
	"github.com/baking-bad/bcdhub/internal/postgres"
	"github.com/baking-bad/bcdhub/internal/postgres/core"
	"github.com/baking-bad/bcdhub/internal/testsuite"
	"github.com/shopspring/decimal"
)

func (s *StorageTestSuite) TestDeleteAll() {
	saver, err := postgres.NewRollback(s.storage.DB)
	s.Require().NoError(err)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	count, err := saver.DeleteAll(ctx, (*block.Block)(nil), 47)
	s.Require().NoError(err)

	err = saver.Commit()
	s.Require().NoError(err)

	s.Require().EqualValues(1, count)

	var block block.Block
	err = s.storage.DB.NewSelect().Model(&block).Order("id desc").Limit(1).Scan(ctx)
	s.Require().NoError(err)

	s.Require().EqualValues(46, block.Level)
}

func (s *StorageTestSuite) TestStatesChangedAtLevel() {
	saver, err := postgres.NewRollback(s.storage.DB)
	s.Require().NoError(err)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	diff, err := saver.StatesChangedAtLevel(ctx, 40)
	s.Require().NoError(err)
	s.Require().Len(diff, 6)

	err = saver.Commit()
	s.Require().NoError(err)
}

func (s *StorageTestSuite) TestLastDiff() {
	saver, err := postgres.NewRollback(s.storage.DB)
	s.Require().NoError(err)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	diff, err := saver.LastDiff(ctx, 41, "exprurUjYU5axnk1qjE6F2t7uDtqR64bnsxGu3AHfWiVREftRDcRPX", true)
	s.Require().NoError(err)

	s.Require().EqualValues(54, diff.ID)
	s.Require().EqualValues(41, diff.Ptr)
	s.Require().EqualValues(40, diff.Level)
	s.Require().EqualValues(3, diff.ProtocolID)
	s.Require().EqualValues(109, diff.OperationID)
	s.Require().EqualValues("KT1NSpRTVR4MUwx64XCADXDUmpMGQw5yVNK1", diff.Contract)
	s.Require().Equal("exprurUjYU5axnk1qjE6F2t7uDtqR64bnsxGu3AHfWiVREftRDcRPX", diff.KeyHash)

	err = saver.Commit()
	s.Require().NoError(err)
}

func (s *StorageTestSuite) TestDeleteBigMapState() {
	saver, err := postgres.NewRollback(s.storage.DB)
	s.Require().NoError(err)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	err = saver.DeleteBigMapState(ctx, bigmapdiff.BigMapState{ID: 54})
	s.Require().NoError(err)

	err = saver.Commit()
	s.Require().NoError(err)

	var state bigmapdiff.BigMapState
	err = s.storage.DB.NewSelect().Model(&state).Where("id = 54").Scan(ctx)
	s.Require().Error(err)
}

func (s *StorageTestSuite) TestSaveBigMapState() {
	saver, err := postgres.NewRollback(s.storage.DB)
	s.Require().NoError(err)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	ts := time.Now().UTC()

	err = saver.SaveBigMapState(ctx, bigmapdiff.BigMapState{
		ID:              54,
		LastUpdateLevel: 1000,
		LastUpdateTime:  ts,
		Value:           types.MustNewBytes("1122"),
		Removed:         true,
		Key:             types.MustNewBytes("7b226279746573223a223030303062333932376530353637626539643736396362666336376564653138613166303430313135313336227d"),
	})
	s.Require().NoError(err)

	err = saver.Commit()
	s.Require().NoError(err)

	var state bigmapdiff.BigMapState
	err = s.storage.DB.NewSelect().Model(&state).Where("id = 54").Scan(ctx)
	s.Require().NoError(err)
	s.Require().EqualValues(1000, state.LastUpdateLevel)
	s.Require().EqualValues(true, state.Removed)
	s.Require().Equal([]byte{0x11, 0x22}, []byte(state.Value))
	s.Require().Equal(ts.Format(time.RFC3339), state.LastUpdateTime.Format(time.RFC3339))
}

func (s *StorageTestSuite) TestGetOperations() {
	saver, err := postgres.NewRollback(s.storage.DB)
	s.Require().NoError(err)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	ops, err := saver.GetOperations(ctx, 40)
	s.Require().NoError(err)

	err = saver.Commit()
	s.Require().NoError(err)

	s.Require().Len(ops, 13)
}

func (s *StorageTestSuite) TestGetLastAction() {
	saver, err := postgres.NewRollback(s.storage.DB)
	s.Require().NoError(err)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	actions, err := saver.GetLastAction(ctx, 43, 46)
	s.Require().NoError(err)

	err = saver.Commit()
	s.Require().NoError(err)

	s.Require().Len(actions, 2)

	s.Require().EqualValues(43, actions[0].AccountId)
	s.Require().EqualValues("2022-01-25T16:45:09Z", actions[0].Time.Format(time.RFC3339))
	s.Require().EqualValues(46, actions[1].AccountId)
	s.Require().EqualValues("2022-01-25T16:45:09Z", actions[1].Time.Format(time.RFC3339))
}

func (s *StorageTestSuite) TestUpdateAccountStats() {
	saver, err := postgres.NewRollback(s.storage.DB)
	s.Require().NoError(err)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	ts := time.Now().UTC()

	err = saver.UpdateAccountStats(ctx, account.Account{
		ID:              43,
		LastAction:      ts,
		OperationsCount: 1,
	})
	s.Require().NoError(err)

	err = saver.Commit()
	s.Require().NoError(err)

	var acc account.Account
	err = s.storage.DB.NewSelect().Model(&acc).Where("id = 43").Scan(ctx)
	s.Require().NoError(err)
	s.Require().EqualValues(1, acc.OperationsCount)
	s.Require().EqualValues(ts.Format(time.RFC3339), acc.LastAction.Format(time.RFC3339))
}

func (s *StorageTestSuite) TestProtocols() {
	saver, err := postgres.NewRollback(s.storage.DB)
	s.Require().NoError(err)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	err = saver.Protocols(ctx, 2)
	s.Require().NoError(err)

	err = saver.Commit()
	s.Require().NoError(err)

	var proto protocol.Protocol
	err = s.storage.DB.NewSelect().Model(&proto).Order("id desc").Limit(1).Scan(ctx)
	s.Require().NoError(err)
	s.Require().EqualValues(0, proto.EndLevel)
	s.Require().EqualValues("Ps9mPmXaRzmzk35gbAYNCAw6UXdE2qoABTHbN2oEEc1qM7CwT9P", proto.Hash)
}

// TestProtocolsRemigrationAfterRollback rolls back the activation block of the
// second protocol and migrates to the same protocol again, as the indexer does
// when the activation block is reorged: the reloaded state must point to the
// first protocol so the migration is not skipped, and after it both protocol
// rows and the new block must be consistent.
func (s *StorageTestSuite) TestProtocolsRemigrationAfterRollback() {
	const (
		activationLevel = 2
		p1Hash          = "Ps9mPmXaRzmzk35gbAYNCAw6UXdE2qoABTHbN2oEEc1qM7CwT9P"
		p2Hash          = "PtHangz2aRngywmSRGGvrcTyMbbdpWdpFKuS4uMWxg2RaH9i1qx"
	)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	head, err := s.blocks.Last(ctx)
	s.Require().NoError(err)
	s.Require().Greater(head.Level, int64(activationLevel))

	saver, err := postgres.NewRollback(s.storage.DB)
	s.Require().NoError(err)
	for level := head.Level; level >= activationLevel; level-- {
		_, err := saver.DeleteAll(ctx, (*block.Block)(nil), level)
		s.Require().NoError(err)
		s.Require().NoError(saver.Protocols(ctx, level))
	}
	s.Require().NoError(saver.Commit())

	// the indexer reloads its state and protocol after the rollback
	state, err := s.blocks.Last(ctx)
	s.Require().NoError(err)
	s.Require().EqualValues(activationLevel-1, state.Level)

	current, err := s.protocols.Get(ctx, "", state.Level)
	s.Require().NoError(err)
	s.Require().Equal(p1Hash, current.Hash)
	s.Require().EqualValues(0, current.EndLevel)

	_, err = s.protocols.Get(ctx, p2Hash, -1)
	s.Require().Error(err, "the second protocol must be deleted by the rollback")
	s.Require().True(s.storage.IsRecordNotFound(err))

	// the new activation block brings the same protocol: migrate again
	tx, err := core.NewTransaction(ctx, s.storage.DB)
	s.Require().NoError(err)
	defer tx.Close()

	current.EndLevel = activationLevel - 1
	s.Require().NoError(tx.Protocol(ctx, &current))

	newProto := protocol.Protocol{
		Hash:       p2Hash,
		StartLevel: activationLevel,
		SymLink:    "babylon",
		Alias:      "PtHangz2",
		ChainID:    current.ChainID,
		Constants:  &protocol.Constants{TimeBetweenBlocks: 15},
	}
	s.Require().NoError(tx.Protocol(ctx, &newProto))
	s.Require().NotZero(newProto.ID)

	newBlock := block.Block{
		Hash:       "BLnewActivationBlock",
		Level:      activationLevel,
		Timestamp:  time.Date(2022, 1, 25, 15, 10, 11, 0, time.UTC),
		ProtocolID: newProto.ID,
	}
	s.Require().NoError(tx.Block(ctx, &newBlock))
	s.Require().NoError(tx.Commit())

	p1, err := s.protocols.Get(ctx, p1Hash, -1)
	s.Require().NoError(err)
	s.Require().EqualValues(activationLevel-1, p1.EndLevel)

	p2, err := s.protocols.Get(ctx, p2Hash, -1)
	s.Require().NoError(err)
	s.Require().Equal(newProto.ID, p2.ID)
	s.Require().EqualValues(activationLevel, p2.StartLevel)
	s.Require().EqualValues(0, p2.EndLevel)

	last, err := s.blocks.Last(ctx)
	s.Require().NoError(err)
	s.Require().EqualValues(activationLevel, last.Level)
	s.Require().Equal(p2.ID, last.ProtocolID)

	byID, err := s.protocols.GetByID(ctx, last.ProtocolID)
	s.Require().NoError(err)
	s.Require().Equal(p2Hash, byID.Hash)
}

func (s *StorageTestSuite) TestRollbackUpdateStats() {
	saver, err := postgres.NewRollback(s.storage.DB)
	s.Require().NoError(err)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	err = saver.UpdateStats(ctx, stats.Stats{
		ID:                1,
		ContractsCount:    1,
		OperationsCount:   188,
		OriginationsCount: 119,
		TransactionsCount: 71,
		EventsCount:       1,
	})
	s.Require().NoError(err)

	err = saver.Commit()
	s.Require().NoError(err)

	var stats stats.Stats
	err = s.storage.DB.NewSelect().Model(&stats).Limit(1).Scan(ctx)
	s.Require().NoError(err)
	s.Require().EqualValues(1, stats.ContractsCount)
	s.Require().EqualValues(188, stats.OperationsCount)
	s.Require().EqualValues(71, stats.TransactionsCount)
	s.Require().EqualValues(119, stats.OriginationsCount)
	s.Require().EqualValues(1, stats.EventsCount)
	s.Require().EqualValues(0, stats.SrOriginationsCount)
}

func (s *StorageTestSuite) TestGetMigrations() {
	saver, err := postgres.NewRollback(s.storage.DB)
	s.Require().NoError(err)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	migrations, err := saver.GetMigrations(ctx, 2)
	s.Require().NoError(err)

	err = saver.Commit()
	s.Require().NoError(err)
	s.Require().Len(migrations, 3)

	migration := migrations[0]
	s.Require().EqualValues(1, migration.Contract.AccountID)
}

func (s *StorageTestSuite) TestGetTicketUpdates() {
	saver, err := postgres.NewRollback(s.storage.DB)
	s.Require().NoError(err)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	updates, err := saver.GetTicketUpdates(ctx, 40)
	s.Require().NoError(err)

	err = saver.Commit()
	s.Require().NoError(err)
	s.Require().Len(updates, 3)

	update := updates[0]
	s.Require().EqualValues(105, update.AccountId)
	s.Require().EqualValues(133, update.Ticket.TicketerID)
}

func (s *StorageTestSuite) TestUpdateTicket() {
	saver, err := postgres.NewRollback(s.storage.DB)
	s.Require().NoError(err)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	err = saver.UpdateTicket(ctx, ticket.Ticket{
		ID:           1,
		ContentType:  testsuite.MustHexDecode("7b227072696d223a22737472696e67227d"),
		Content:      testsuite.MustHexDecode("7b22737472696e67223a22616263227d"),
		TicketerID:   133,
		UpdatesCount: 1,
	})
	s.Require().NoError(err)

	err = saver.Commit()
	s.Require().NoError(err)

	var t ticket.Ticket
	err = s.storage.DB.NewSelect().Model(&t).Where("id = 1").Scan(ctx)
	s.Require().NoError(err)
	s.Require().EqualValues(133, t.TicketerID)
	s.Require().EqualValues(1, t.UpdatesCount)
}

func (s *StorageTestSuite) TestGetLastActionNoRows() {
	saver, err := postgres.NewRollback(s.storage.DB)
	s.Require().NoError(err)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	actions, err := saver.GetLastAction(ctx, 1000000)
	s.Require().NoError(err)

	err = saver.Commit()
	s.Require().NoError(err)

	s.Require().Len(actions, 0)
}

func (s *StorageTestSuite) TestTicketBalancesRollback() {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	saver, err := postgres.NewRollback(s.storage.DB)
	s.Require().NoError(err)

	err = saver.TicketBalances(ctx, &ticket.Balance{
		TicketId:  1,
		AccountId: 131,
		Amount:    decimal.RequireFromString("43"),
	})
	s.Require().NoError(err)

	err = saver.Commit()
	s.Require().NoError(err)

	var balances []ticket.Balance
	err = s.storage.DB.NewSelect().Model(&balances).Where("amount > 0").Where("ticket_id = 1").Scan(ctx)
	s.Require().NoError(err)
	s.Require().Len(balances, 1)
}

func (s *StorageTestSuite) TestDeleteTickets() {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	saver, err := postgres.NewRollback(s.storage.DB)
	s.Require().NoError(err)

	ids, err := saver.DeleteTickets(ctx, 40)
	s.Require().NoError(err)

	err = saver.Commit()
	s.Require().NoError(err)

	s.Require().Len(ids, 2)
	s.Require().ElementsMatch([]int64{1, 2}, ids)
}

func (s *StorageTestSuite) TestDeleteTicketBalances() {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	saver, err := postgres.NewRollback(s.storage.DB)
	s.Require().NoError(err)

	err = saver.DeleteTicketBalances(ctx, []int64{1})
	s.Require().NoError(err)

	err = saver.Commit()
	s.Require().NoError(err)

	var balances []ticket.Balance
	err = s.storage.DB.NewSelect().Model(&balances).Where("ticket_id = 1").Scan(ctx)
	s.Require().NoError(err)
	s.Require().Len(balances, 0)
}
