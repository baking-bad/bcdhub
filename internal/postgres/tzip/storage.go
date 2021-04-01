package tzip

import (
	"github.com/baking-bad/bcdhub/internal/models"
	"github.com/baking-bad/bcdhub/internal/models/tzip"
	"github.com/baking-bad/bcdhub/internal/postgres/core"
)

// Storage -
type Storage struct {
	*core.Postgres
}

// NewStorage -
func NewStorage(pg *core.Postgres) *Storage {
	return &Storage{pg}
}

// Get -
func (storage *Storage) Get(network, address string) (*tzip.TZIP, error) {
	var t tzip.TZIP
	err := storage.DB.
		Table(models.DocTZIP).
		Scopes(core.NetworkAndAddress(network, address)).
		First(&t).
		Error
	return &t, err
}

// GetBySlug -
func (storage *Storage) GetBySlug(slug string) (*tzip.TZIP, error) {
	var t tzip.TZIP
	err := storage.DB.
		Table(models.DocTZIP).
		Where("slug = ?", slug).
		First(&t).
		Error
	return &t, err
}

// GetAliasesMap -
func (storage *Storage) GetAliasesMap(network string) (map[string]string, error) {
	var t []tzip.TZIP
	if err := storage.DB.
		Table(models.DocTZIP).
		Where("network = ?", network).
		Find(&t).Error; err != nil {
		return nil, err
	}

	aliases := make(map[string]string)
	for i := range t {
		aliases[t[i].Address] = t[i].Name
	}

	return aliases, nil
}

// GetAliases -
func (storage *Storage) GetAliases(network string) (t []tzip.TZIP, err error) {
	err = storage.DB.
		Table(models.DocTZIP).
		Where("network = ?", network).
		Where("name IS NOT NULL").
		Find(&t).Error

	return
}

// GetWithEvents -
func (storage *Storage) GetWithEvents() (t []tzip.TZIP, err error) {
	err = storage.DB.
		Table(models.DocTZIP).
		Where("events is not null AND jsonb_array_length(events) > 0").
		Find(&t).Error
	return
}

// GetWithEventsCounts -
func (storage *Storage) GetWithEventsCounts() (int64, error) {
	var count int64
	err := storage.DB.
		Table(models.DocTZIP).
		Where("events is not null AND jsonb_array_length(events) > 0").
		Count(&count).
		Error
	return count, err
}
