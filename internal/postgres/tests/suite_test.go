package tests

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/baking-bad/bcdhub/internal/postgres"
	"github.com/baking-bad/bcdhub/internal/postgres/account"
	"github.com/baking-bad/bcdhub/internal/postgres/bigmapaction"
	"github.com/baking-bad/bcdhub/internal/postgres/bigmapdiff"
	"github.com/baking-bad/bcdhub/internal/postgres/block"
	"github.com/baking-bad/bcdhub/internal/postgres/contract"
	"github.com/baking-bad/bcdhub/internal/postgres/core"
	"github.com/baking-bad/bcdhub/internal/postgres/domains"
	"github.com/baking-bad/bcdhub/internal/postgres/global_constant"
	"github.com/baking-bad/bcdhub/internal/postgres/migration"
	"github.com/baking-bad/bcdhub/internal/postgres/operation"
	"github.com/baking-bad/bcdhub/internal/postgres/protocol"
	smartrollup "github.com/baking-bad/bcdhub/internal/postgres/smart_rollup"
	"github.com/baking-bad/bcdhub/internal/postgres/ticket"
	"github.com/dipdup-net/go-lib/database"
	"github.com/go-testfixtures/testfixtures/v3"
	"github.com/stretchr/testify/suite"
)

// StorageTestSuite -
type StorageTestSuite struct {
	suite.Suite
	psqlContainer *database.PostgreSQLContainer
	storage       *core.Postgres

	accounts        *account.Storage
	bigMapActions   *bigmapaction.Storage
	bigMapDiffs     *bigmapdiff.Storage
	blocks          *block.Storage
	contracts       *contract.Storage
	domains         *domains.Storage
	globalConstants *global_constant.Storage
	migrations      *migration.Storage
	operations      *operation.Storage
	protocols       *protocol.Storage
	smartRollups    *smartrollup.Storage
	ticketUpdates   *ticket.Storage
}

// SetupSuite -
func (s *StorageTestSuite) SetupSuite() {
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

	db, err := sql.Open("postgres", psqlContainer.GetDSN())
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

	s.accounts = account.NewStorage(strg)
	s.bigMapActions = bigmapaction.NewStorage(strg)
	s.bigMapDiffs = bigmapdiff.NewStorage(strg)
	s.blocks = block.NewStorage(strg)
	s.contracts = contract.NewStorage(strg)
	s.domains = domains.NewStorage(strg)
	s.globalConstants = global_constant.NewStorage(strg)
	s.migrations = migration.NewStorage(strg)
	s.operations = operation.NewStorage(strg)
	s.protocols = protocol.NewStorage(strg)
	s.smartRollups = smartrollup.NewStorage(strg)
	s.ticketUpdates = ticket.NewStorage(strg)
}

// TearDownSuite -
func (s *StorageTestSuite) TearDownSuite() {
	ctx, ctxCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer ctxCancel()

	s.Require().NoError(s.storage.Close())
	s.Require().NoError(s.psqlContainer.Terminate(ctx))
}

func TestSuiteStorage_Run(t *testing.T) {
	suite.Run(t, new(StorageTestSuite))
}
