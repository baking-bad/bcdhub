package core

import (
	"github.com/baking-bad/bcdhub/internal/models"
)

// GetByID -
func (p *Postgres) GetByID(output models.Model) error {
	err := p.DB.Table(output.GetIndex()).First(output, output.GetID()).Error
	return err
}

// GetAll -
func (p *Postgres) GetAll(index string) ([]models.Model, error) {
	var result []models.Model
	err := p.DB.Table(index).Find(&result).Error
	return result, err
}

// GetByIDs -
func (p *Postgres) GetByIDs(index string, ids ...int64) ([]models.Model, error) {
	var result []models.Model
	err := p.DB.Table(index).Find(&result, ids).Error
	return result, err
}

// GetByNetwork -
func (p *Postgres) GetByNetwork(network, index string) ([]models.Model, error) {
	var result []models.Model
	err := p.DB.Table(index).Where("network = ?", network).Find(&result).Error
	return result, err
}
