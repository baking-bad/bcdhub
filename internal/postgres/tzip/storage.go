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
	t := new(tzip.TZIP)
	query := storage.DB.Model(t)
	core.NetworkAndAddress(network, address)(query)
	err := query.Order("id desc").Limit(1).Select()
	return t, err
}

// GetBySlug -
func (storage *Storage) GetBySlug(slug string) (*tzip.TZIP, error) {
	t := new(tzip.TZIP)
	err := storage.DB.Model(t).Where("slug = ?", slug).First()
	return t, err
}

// GetAliases -
func (storage *Storage) GetAliases(network types.Network) (t []tzip.TZIP, err error) {
	err = storage.DB.Model(&t).
		Where("network = ?", network).
		Where("name IS NOT NULL").
		Select(&t)

	return
}

// GetWithEvents -
func (storage *Storage) GetWithEvents(updatedAt uint64) ([]tzip.TZIP, error) {
	query := storage.DB.Model().
		Table(models.DocTZIP).
		Where("events is not null AND jsonb_array_length(events) > 0")

	if updatedAt > 0 {
		query.Where("updated_at > ?", updatedAt)
	}

	t := make([]tzip.TZIP, 0)
	if err := query.Order("updated_at asc").Select(&t); err != nil && !storage.IsRecordNotFound(err) {
		return nil, err
	}
	return t, nil
}
