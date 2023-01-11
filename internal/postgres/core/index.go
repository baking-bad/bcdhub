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
