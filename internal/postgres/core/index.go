package core

import (
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

// DeleteTables -
func (p *Postgres) DeleteTables(indices []string) error {
	for _, index := range indices {
		if err := p.DB.Model().Table(index).DropTable(&orm.DropTableOptions{
			IfExists: true,
		}); err != nil {
			return err
		}
	}
	return nil
}
