package tests

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/baking-bad/bcdhub/internal/models/account"
	"github.com/baking-bad/bcdhub/internal/models/bigmapdiff"
	"github.com/baking-bad/bcdhub/internal/models/block"
	"github.com/baking-bad/bcdhub/internal/models/protocol"
	"github.com/baking-bad/bcdhub/internal/models/types"
	"github.com/baking-bad/bcdhub/internal/postgres"
	"github.com/baking-bad/bcdhub/internal/postgres/core"
	"github.com/dipdup-net/go-lib/database"
	"github.com/go-testfixtures/testfixtures/v3"
	"github.com/stretchr/testify/suite"
)

// RollbackTestSuite -
type RollbackTestSuite struct {
	suite.Suite
	psqlContainer *database.PostgreSQLContainer
	storage       *core.Postgres
}

// SetupSuite -
func (s *RollbackTestSuite) SetupSuite() {
	ctx, ctxCancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer ctxCancel()

	psqlContainer, err := database.NewPostgreSQLContainer(ctx, database.PostgreSQLContainerConfig{
		User:     "user",
		Password: "password",
		Database: "db_test",
		Port:     5432,
		Image:    "postgres:14",
	})
	s.Require().NoError(err)
	s.psqlContainer = psqlContainer

	strg, err := core.New(core.Config{
		User:     s.psqlContainer.Config.User,
		DBName:   s.psqlContainer.Config.Database,
		Password: s.psqlContainer.Config.Password,
		Host:     s.psqlContainer.Config.Host,
		Port:     s.psqlContainer.MappedPort().Int(),
		SslMode:  "disable",
	}, "public", "bcd")
	s.Require().NoError(err)
	s.storage = strg

	err = strg.InitDatabase(ctx)
	s.Require().NoError(err)

	pm := postgres.NewPartitionManager(strg)
	err = pm.CreatePartitions(ctx, time.Date(2022, 1, 1, 1, 1, 1, 1, time.Local))
	s.Require().NoError(err)
}

// TearDownSuite -
func (s *RollbackTestSuite) TearDownSuite() {
	ctx, ctxCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer ctxCancel()

	s.Require().NoError(s.storage.Close())
	s.Require().NoError(s.psqlContainer.Terminate(ctx))
}

func (s *RollbackTestSuite) SetupTest() {
	db, err := sql.Open("postgres", s.psqlContainer.GetDSN())
	s.Require().NoError(err)

	fixtures, err := testfixtures.New(
		testfixtures.Database(db),
		testfixtures.Dialect("postgres"),
		testfixtures.Directory("./fixtures"),
		testfixtures.UseAlterConstraint(),
	)
	s.Require().NoError(err)
	s.Require().NoError(fixtures.Load())
	s.Require().NoError(db.Close())
}

func TestSuiteRollback_Run(t *testing.T) {
	suite.Run(t, new(RollbackTestSuite))
}

func (s *RollbackTestSuite) TestDeleteAll() {
	saver, err := postgres.NewRollback(s.storage.DB)
	s.Require().NoError(err)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	err = saver.DeleteAll(ctx, (*block.Block)(nil), 47)
	s.Require().NoError(err)

	err = saver.Commit()
	s.Require().NoError(err)

	var block block.Block
	err = s.storage.DB.NewSelect().Model(&block).Order("id desc").Limit(1).Scan(ctx)
	s.Require().NoError(err)

	s.Require().EqualValues(46, block.Level)
}

func (s *RollbackTestSuite) TestStatesChangedAtLevel() {
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

func (s *RollbackTestSuite) TestLastDiff() {
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

func (s *RollbackTestSuite) TestDeleteBigMapState() {
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

func (s *RollbackTestSuite) TestSaveBigMapState() {
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

func (s *RollbackTestSuite) TestGetOperations() {
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

func (s *RollbackTestSuite) TestGetLastAction() {
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

func (s *RollbackTestSuite) TestUpdateAccountStats() {
	saver, err := postgres.NewRollback(s.storage.DB)
	s.Require().NoError(err)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	ts := time.Now().UTC()

	err = saver.UpdateAccountStats(ctx, 43, ts, 1)
	s.Require().NoError(err)

	err = saver.Commit()
	s.Require().NoError(err)

	var acc account.Account
	err = s.storage.DB.NewSelect().Model(&acc).Where("id = 43").Scan(ctx)
	s.Require().NoError(err)
	s.Require().EqualValues(1, acc.OperationsCount)
	s.Require().EqualValues(ts.Format(time.RFC3339), acc.LastAction.Format(time.RFC3339))
}

func (s *RollbackTestSuite) TestProtocols() {
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
