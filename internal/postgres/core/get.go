package core

import (
	"github.com/baking-bad/bcdhub/internal/models"
	"github.com/go-pg/pg/v10"
)

// GetByID -
func (p *Postgres) GetByID(output models.Model) error {
	err := p.DB.Model().Table(output.GetIndex()).Where("id = ?", output.GetID()).Select(output)
	return err
}

// GetAll -
func (p *Postgres) GetAll(index string) ([]models.Model, error) {
	var result []models.Model
	err := p.DB.Model().Table(index).Select(&result)
	return result, err
}

// GetByIDs -
func (p *Postgres) GetByIDs(index string, ids ...int64) ([]models.Model, error) {
	var result []models.Model
	err := p.DB.Model().Table(index).Where("id IN (?)", pg.In(ids)).Select(&result)
	return result, err
}
