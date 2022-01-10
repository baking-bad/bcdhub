package contract_metadata

import (
	cm "github.com/baking-bad/bcdhub/internal/models/contract_metadata"
	"github.com/baking-bad/bcdhub/internal/models/types"
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
func (storage *Storage) Get(network types.Network, address string) (*cm.ContractMetadata, error) {
	t := new(cm.ContractMetadata)
	query := storage.DB.Model(t)
	core.NetworkAndAddress(network, address)(query)
	err := query.Order("id desc").Limit(1).Select()
	return t, err
}

// GetBySlug -
func (storage *Storage) GetBySlug(slug string) (*cm.ContractMetadata, error) {
	t := new(cm.ContractMetadata)
	err := storage.DB.Model(t).Where("slug = ?", slug).First()
	return t, err
}

// GetAliases -
func (storage *Storage) GetAliases(network types.Network) (t []cm.ContractMetadata, err error) {
	err = storage.DB.Model(&t).
		Where("network = ?", network).
		Where("name IS NOT NULL").
		Select(&t)

	return
}

// GetWithEvents -
func (storage *Storage) GetWithEvents(updatedAt uint64) ([]cm.ContractMetadata, error) {
	query := storage.DB.Model((*cm.ContractMetadata)(nil)).
		Where("events is not null AND jsonb_array_length(events) > 0")

	if updatedAt > 0 {
		query.Where("updated_at > ?", updatedAt)
	}

	t := make([]cm.ContractMetadata, 0)
	if err := query.Order("updated_at asc").Select(&t); err != nil && !storage.IsRecordNotFound(err) {
		return nil, err
	}
	return t, nil
}

// Events -
func (storage *Storage) Events(network types.Network, address string) (events cm.Events, err error) {
	err = storage.DB.Model((*cm.ContractMetadata)(nil)).Column("events").
		Where("network = ?", network).Where("address = ?", address).
		Limit(1).Select(&events)
	return
}
