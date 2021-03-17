package core

import (
	"github.com/baking-bad/bcdhub/internal/models"
)

// CreateIndexes -
func (p *Postgres) CreateIndexes() error {
	for _, index := range models.AllModels() {
		if p.DB.Migrator().HasTable(index) {
			continue
		}

		if err := p.DB.Migrator().CreateTable(index); err != nil {
			return err
		}
	}
	return nil
}

// DeleteIndices -
func (p *Postgres) DeleteIndices(indices []string) error {
	for _, index := range indices {
		if err := p.DB.Migrator().DropTable(index); err != nil {
			return err
		}
	}
	return nil
}
