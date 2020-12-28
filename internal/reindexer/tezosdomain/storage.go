package tezosdomain

import (
	"github.com/baking-bad/bcdhub/internal/helpers"
	"github.com/baking-bad/bcdhub/internal/models"
	"github.com/baking-bad/bcdhub/internal/models/tezosdomain"
	"github.com/baking-bad/bcdhub/internal/reindexer/core"
	"github.com/pkg/errors"
)

// Storage -
type Storage struct {
	db *core.Reindexer
}

// NewStorage -
func NewStorage(db *core.Reindexer) *Storage {
	return &Storage{db}
}

// ListDomains -
func (storage *Storage) ListDomains(network string, size, offset int64) (tezosdomain.DomainsResponse, error) {
	if size > core.DefaultSize {
		size = core.DefaultSize
	}

	query := storage.db.Query(models.DocTezosDomains).
		Match("network", network).
		Sort("timestamp", true).
		Limit(int(size)).
		Offset(int(offset))

	var total int
	domains := make([]tezosdomain.TezosDomain, 0)
	if err := storage.db.GetAllByQueryWithTotal(query, &total, &domains); err != nil {
		return tezosdomain.DomainsResponse{}, nil
	}

	return tezosdomain.DomainsResponse{
		Domains: domains,
		Total:   int64(total),
	}, nil
}

// ResolveDomainByAddress -
func (storage *Storage) ResolveDomainByAddress(network string, address string) (*tezosdomain.TezosDomain, error) {
	if !helpers.IsAddress(address) {
		return nil, errors.Errorf("Invalid address: %s", address)
	}

	query := storage.db.Query(models.DocTezosDomains).
		Match("network", network).
		Match("address", address)

	var td tezosdomain.TezosDomain
	err := storage.db.GetOne(query, &td)
	return &td, err
}
