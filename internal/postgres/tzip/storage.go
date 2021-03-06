package tzip

import (
	"github.com/baking-bad/bcdhub/internal/models"
	"github.com/baking-bad/bcdhub/internal/models/types"
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
func (storage *Storage) Get(network types.Network, address string) (*tzip.TZIP, error) {
	var t tzip.TZIP
	err := storage.DB.
		Table(models.DocTZIP).
		Scopes(core.NetworkAndAddress(network, address)).
		Order("id desc").
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

// GetAliases -
func (storage *Storage) GetAliases(network types.Network) (t []tzip.TZIP, err error) {
	err = storage.DB.
		Table(models.DocTZIP).
		Where("network = ?", network).
		Where("name IS NOT NULL").
		Find(&t).Error

	return
}

// GetWithEvents -
func (storage *Storage) GetWithEvents(updatedAt uint64) ([]tzip.TZIP, error) {
	query := storage.DB.
		Table(models.DocTZIP).
		Where("events is not null AND jsonb_array_length(events) > 0")

	if updatedAt > 0 {
		query.Where("updated_at > ?", updatedAt)
	}

	t := make([]tzip.TZIP, 0)
	if err := query.Order("updated_at asc").Find(&t).Error; err != nil && !storage.IsRecordNotFound(err) {
		return nil, err
	}
	return t, nil
}
