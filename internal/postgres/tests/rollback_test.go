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
	"github.com/baking-bad/bcdhub/internal/testsuite"
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
	s.Require().EqualValues(105, update.AccountID)
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
