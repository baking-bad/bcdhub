package core

import (
	"context"

	"github.com/baking-bad/bcdhub/internal/models"
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

	if by := model.PartitionBy(); by != "" {
		query = query.PartitionBy(by)
	}

	_, err := query.Exec(ctx)
	return err
}

func createTables(ctx context.Context, db *bun.DB) error {
	// register many-to-many relationships
	db.RegisterModel(models.ManyToMany()...)

	for _, model := range models.AllModels() {
		if err := createTable(ctx, db, model); err != nil {
			return err
		}
	}
	return nil
}

func createSchema(db *bun.DB, schemaName string) error {
	schema := bun.Ident(schemaName)
	if _, err := db.Exec("create schema if not exists ?", schema); err != nil {
		return err
	}
	_, err := db.Exec("set search_path = ?", schema)
	return err
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
		err := p.DB.QueryRow(tableExistsQuery, p.schema, table).Scan(&exists)
		if !exists.Flag || err != nil {
			return false
		}
	}
	return true
}
