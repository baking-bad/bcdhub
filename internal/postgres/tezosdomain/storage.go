package tezosdomain

import (
	"github.com/baking-bad/bcdhub/internal/helpers"
	"github.com/baking-bad/bcdhub/internal/models"
	"github.com/baking-bad/bcdhub/internal/models/tezosdomain"
	"github.com/baking-bad/bcdhub/internal/postgres/consts"
	"github.com/baking-bad/bcdhub/internal/postgres/core"
	"github.com/pkg/errors"
)

// Storage -
type Storage struct {
	*core.Postgres
}

// NewStorage -
func NewStorage(pg *core.Postgres) *Storage {
	return &Storage{pg}
}

// ListDomains -
func (storage *Storage) ListDomains(network string, size, offset int64) (tezosdomain.DomainsResponse, error) {
	limit := core.GetPageSize(size)

	response := tezosdomain.DomainsResponse{
		Domains: make([]tezosdomain.TezosDomain, 0),
	}

	err := storage.DB.Table(models.DocTezosDomains).
		Where("network = ?", network).
		Limit(limit).
		Offset(int(offset)).
		Find(&response.Domains).
		Error
	if err != nil {
		return response, err
	}

	err = storage.DB.Table(models.DocTezosDomains).
		Where("network = ?", network).
		Count(&response.Total).
		Error

	return response, err
}

// ResolveDomainByAddress -
func (storage *Storage) ResolveDomainByAddress(network string, address string) (*tezosdomain.TezosDomain, error) {
	if !helpers.IsAddress(address) {
		return nil, errors.Wrapf(consts.ErrInvalidAddress, "ResolveDomainByAddress %s", address)
	}

	var td tezosdomain.TezosDomain
	err := storage.DB.
		Table(models.DocTezosDomains).
		Scopes(core.NetworkAndAddress(network, address)).
		First(&td).
		Error

	return &td, err
}
