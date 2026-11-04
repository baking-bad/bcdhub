package core

import (
	"context"
	"database/sql"
	"errors"
	"sync"

	"github.com/baking-bad/bcdhub/internal/models"
	"github.com/baking-bad/bcdhub/internal/models/bigmapaction"
	"github.com/baking-bad/bcdhub/internal/models/bigmapdiff"
	"github.com/baking-bad/bcdhub/internal/models/block"
	"github.com/baking-bad/bcdhub/internal/models/contract"
	"github.com/baking-bad/bcdhub/internal/models/migration"
	"github.com/baking-bad/bcdhub/internal/models/operation"
	"github.com/baking-bad/bcdhub/internal/models/ticket"
	"github.com/rs/zerolog/log"
	"github.com/uptrace/bun"
)

func createTable(ctx context.Context, db bun.IDB, model models.Model) error {
	if model == nil {
		return nil
	}

	query := db.
		NewCreateTable().
		Model(model).
		IfNotExists()

	_, err := query.Exec(ctx)
	return err
}

func createTables(ctx context.Context, db *bun.DB) error {
	for _, model := range models.AllModels() {
		if err := createTable(ctx, db, model); err != nil {
			return err
		}
	}
	return createHypertables(ctx, db)
}

func createHypertables(ctx context.Context, db *bun.DB) error {
	for _, model := range []models.Model{
		&block.Block{},
		&bigmapdiff.BigMapDiff{},
		&bigmapaction.BigMapAction{},
		&contract.Contract{},
		&migration.Migration{},
		&operation.Operation{},
		&ticket.TicketUpdate{},
	} {
		if _, err := db.ExecContext(ctx,
			`SELECT public.create_hypertable(?, 'timestamp', chunk_time_interval => INTERVAL '1 month', if_not_exists => TRUE);`,
			model.TableName(),
		); err != nil {
			return err
		}
	}

	return nil
}

var createExtensionOnce sync.Once
var wg sync.WaitGroup

func createSchema(ctx context.Context, db *bun.DB, schemaName string) error {
	createExtensionOnce.Do(func() {
		defer wg.Done()
		wg.Add(1)
		if _, err := db.NewRaw("set search_path = 'public'").Exec(ctx); err != nil {
			log.Err(err).Msg("set search path to public")
			return
		}
		if _, err := db.NewRaw("CREATE EXTENSION IF NOT EXISTS timescaledb;").Exec(ctx); err != nil {
			log.Err(err).Msg("create timescale extension")
		}
	})

	wg.Wait()

	schema := bun.Ident(schemaName)
	if _, err := db.NewRaw("create schema if not exists ?", schema).Exec(ctx); err != nil {
		return err
	}
	if _, err := db.NewRaw("set search_path = ?", schema).Exec(ctx); err != nil {
		return err
	}
	return nil
}

// Drop - drops full database
func (p *Postgres) Drop(ctx context.Context) error {
	for _, table := range models.ManyToMany() {
		if _, err := p.DB.NewDropTable().Model(table).IfExists().Cascade().Exec(ctx); err != nil {
			return err
		}
	}

	for _, table := range models.AllModels() {
		if _, err := p.DB.NewDropTable().Model(table).IfExists().Cascade().Exec(ctx); err != nil {
			return err
		}
	}
	return nil
}

const tableExistsQuery = `SELECT EXISTS(
    SELECT * 
    FROM information_schema.tables 
    WHERE 
      table_schema = ? AND 
      table_name = ?
) as flag;`

type existsResponse struct {
	Flag bool `bun:"flag"`
}

// TablesExist - returns true if all tables exist otherwise false
func (p *Postgres) TablesExist(ctx context.Context) bool {
	for _, table := range models.AllDocuments() {
		var exists existsResponse
		err := p.DB.NewRaw(tableExistsQuery, p.schema, table).Scan(ctx, &exists)
		if !exists.Flag || err != nil {
			return false
		}
	}
	return true
}

// widenTicketAmounts drops the numeric(200,0) limit from ticket amount columns
// of databases created before it was removed. Ticket amounts are unbounded
// nats: anyone can mint a ticket with a 201-digit amount, and the insert then
// failed on every retry, halting the indexer for good. Removing a numeric
// precision limit does not rewrite the table.
func widenTicketAmounts(ctx context.Context, db *bun.DB) error {
	for _, table := range []string{
		(ticket.Balance{}).TableName(),
		(ticket.TicketUpdate{}).TableName(),
	} {
		var precision sql.NullInt64
		err := db.NewRaw(
			`SELECT numeric_precision FROM information_schema.columns WHERE table_schema = current_schema() AND table_name = ? AND column_name = 'amount'`,
			table,
		).Scan(ctx, &precision)
		switch {
		case errors.Is(err, sql.ErrNoRows):
			continue
		case err != nil:
			return err
		case !precision.Valid:
			continue
		}

		log.Info().Str("table", table).Int64("precision", precision.Int64).Msg("removing precision limit of ticket amounts")
		if _, err := db.NewRaw(`ALTER TABLE ? ALTER COLUMN amount TYPE numeric`, bun.Ident(table)).Exec(ctx); err != nil {
			return err
		}
	}
	return nil
}
