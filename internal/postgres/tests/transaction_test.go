package tests

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/baking-bad/bcdhub/internal/models/account"
	"github.com/baking-bad/bcdhub/internal/models/bigmapdiff"
	"github.com/baking-bad/bcdhub/internal/models/contract"
	"github.com/baking-bad/bcdhub/internal/models/migration"
	"github.com/baking-bad/bcdhub/internal/models/protocol"
	"github.com/baking-bad/bcdhub/internal/models/types"
	"github.com/baking-bad/bcdhub/internal/postgres"
	"github.com/baking-bad/bcdhub/internal/postgres/core"
	"github.com/dipdup-net/go-lib/database"
	"github.com/go-testfixtures/testfixtures/v3"
	"github.com/stretchr/testify/suite"
)

// TransactionTest -
type TransactionTest struct {
	suite.Suite
	psqlContainer *database.PostgreSQLContainer
	storage       *core.Postgres
}

// SetupSuite -
func (s *TransactionTest) SetupSuite() {
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
func (s *TransactionTest) TearDownSuite() {
	ctx, ctxCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer ctxCancel()

	s.Require().NoError(s.storage.Close())
	s.Require().NoError(s.psqlContainer.Terminate(ctx))
}

func TestSuiteTransaction_Run(t *testing.T) {
	suite.Run(t, new(TransactionTest))
}

func (s *TransactionTest) TestSave() {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	tx, err := core.NewTransaction(ctx, s.storage.DB)
	s.Require().NoError(err)

	account := account.Account{
		Address: "address",
		Type:    types.AccountTypeContract,
		Alias:   "alias",
	}
	err = tx.Save(ctx, &account)
	s.Require().NoError(err)
	s.Require().Positive(account.ID)

	err = tx.Commit()
	s.Require().NoError(err)
}

func (s *TransactionTest) TestMigrations() {
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

func (s *TransactionTest) TestProtocol() {
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

func (s *TransactionTest) TestScriptConstants() {
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

	var result []contract.ScriptConstants
	err = s.storage.DB.NewSelect().Model(&result).Scan(ctx)
	s.Require().NoError(err)
	s.Require().Len(result, 3)
}

func (s *TransactionTest) TestScripts() {
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

	var result []contract.Script
	err = s.storage.DB.NewSelect().Model(&result).Scan(ctx)
	s.Require().NoError(err)
	s.Require().Len(result, 3)
}

func (s *TransactionTest) TestScriptsConflict() {
	db, err := sql.Open("postgres", s.psqlContainer.GetDSN())
	s.Require().NoError(err)

	fixtures, err := testfixtures.New(
		testfixtures.Database(db),
		testfixtures.Dialect("postgres"),
		testfixtures.Files(
			"./fixtures/scripts.yml",
		),
		testfixtures.UseAlterConstraint(),
	)
	s.Require().NoError(err)
	s.Require().NoError(fixtures.Load())
	s.Require().NoError(db.Close())

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	tx, err := core.NewTransaction(ctx, s.storage.DB)
	s.Require().NoError(err)

	update := contract.Script{
		Hash: "8436dde35bd56644cd4f40c5f26839cb8f4b51052e415da2b9fadcd9bddcb03e",
	}
	err = tx.Scripts(ctx, &update)
	s.Require().NoError(err)
	s.Require().EqualValues(1, update.ID)

	err = tx.Commit()
	s.Require().NoError(err)
}

func (s *TransactionTest) TestAccounts() {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	tx, err := core.NewTransaction(ctx, s.storage.DB)
	s.Require().NoError(err)

	sc := []*account.Account{
		{
			Address: "address_1",
			Type:    types.AccountTypeContract,
			Alias:   "alias_1",
		}, {
			Address: "address_12",
			Type:    types.AccountTypeSmartRollup,
			Alias:   "alias_12",
		}, {
			Address: "address_2",
			Type:    types.AccountTypeTz,
			Alias:   "alias_2",
		},
	}
	err = tx.Accounts(ctx, sc...)
	s.Require().NoError(err)

	err = tx.Commit()
	s.Require().NoError(err)

	var result []account.Account
	err = s.storage.DB.NewSelect().Model(&result).Scan(ctx)
	s.Require().NoError(err)
	s.Require().Len(result, 3)
}

func (s *TransactionTest) TestBigMapStates() {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	tx, err := core.NewTransaction(ctx, s.storage.DB)
	s.Require().NoError(err)

	sc := []*bigmapdiff.BigMapState{
		{
			Key:             []byte{0, 1, 2, 3},
			KeyHash:         "hash 1",
			Ptr:             1,
			LastUpdateLevel: 100,
			Count:           1,
			Removed:         false,
			Contract:        "contract 1",
		}, {
			Key:             []byte{0, 1, 2, 3, 4},
			KeyHash:         "hash 2",
			Ptr:             1,
			LastUpdateLevel: 100,
			Count:           1,
			Removed:         false,
			Contract:        "contract 2",
		}, {
			Key:             []byte{0, 1, 2, 3, 5},
			KeyHash:         "hash 3",
			Ptr:             1,
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
	err = s.storage.DB.NewSelect().Model(&result).Scan(ctx)
	s.Require().NoError(err)
	s.Require().Len(result, 3)
}

func (s *TransactionTest) TestUpdateContracts() {
	db, err := sql.Open("postgres", s.psqlContainer.GetDSN())
	s.Require().NoError(err)

	fixtures, err := testfixtures.New(
		testfixtures.Database(db),
		testfixtures.Dialect("postgres"),
		testfixtures.Files(
			"./fixtures/accounts.yml",
			"./fixtures/contracts.yml",
		),
		testfixtures.UseAlterConstraint(),
	)
	s.Require().NoError(err)
	s.Require().NoError(fixtures.Load())
	s.Require().NoError(db.Close())

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	tx, err := core.NewTransaction(ctx, s.storage.DB)
	s.Require().NoError(err)

	sc := []*contract.Update{
		{
			AccountID:  1,
			LastAction: time.Now(),
			TxCount:    10,
		}, {
			AccountID:  2,
			LastAction: time.Now(),
			TxCount:    10,
		}, {
			AccountID:  3,
			LastAction: time.Now(),
			TxCount:    10,
		},
	}
	err = tx.UpdateContracts(ctx, sc...)
	s.Require().NoError(err)

	err = tx.Commit()
	s.Require().NoError(err)
}
