package core

import (
	"context"

	"github.com/baking-bad/bcdhub/internal/models"
	"github.com/go-pg/pg/v10/orm"
)

// CreateTables -
func (p *Postgres) CreateTables() error {
	for _, index := range models.AllModels() {
		if err := p.DB.Model(index).CreateTable(&orm.CreateTableOptions{
			IfNotExists: true,
		}); err != nil {
			return err
		}
	}
	return nil
}

// Drop - drops full database
func (p *Postgres) Drop(ctx context.Context) error {
	for _, table := range models.ManyToMany() {
		if err := p.DB.Model(table).DropTable(&orm.DropTableOptions{
			IfExists: true,
			Cascade:  true,
		}); err != nil {
			return err
		}
	}

	for _, table := range models.AllModels() {
		if err := p.DB.Model(table).DropTable(&orm.DropTableOptions{
			IfExists: true,
			Cascade:  true,
		}); err != nil {
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
	Flag bool `pg:"flag,use_zero"`
}

// TablesExist - returns true if all tables exist otherwise false
func (p *Postgres) TablesExist() bool {
	for _, table := range models.AllDocuments() {
		var exists existsResponse
		_, err := p.DB.QueryOne(&exists, tableExistsQuery, p.schema, table)
		if !exists.Flag || err != nil {
			return false
		}
	}
	return true
}
