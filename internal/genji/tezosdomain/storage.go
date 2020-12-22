package tezosdomain

import (
	"github.com/baking-bad/bcdhub/internal/genji/core"
	"github.com/baking-bad/bcdhub/internal/helpers"
	"github.com/baking-bad/bcdhub/internal/models"
	"github.com/baking-bad/bcdhub/internal/models/tezosdomain"
	"github.com/genjidb/genji/document"
	"github.com/pkg/errors"
)

// Storage -
type Storage struct {
	db *core.Genji
}

// NewStorage -
func NewStorage(db *core.Genji) *Storage {
	return &Storage{db}
}

// ListDomains -
func (storage *Storage) ListDomains(network string, size, offset int64) (tezosdomain.DomainsResponse, error) {
	if size > core.DefaultSize {
		size = core.DefaultSize
	}

	builder := core.NewBuilder().SelectAll(models.DocTezosDomains).And(
		core.NewEq("network", network),
	).SortDesc("timestamp").Limit(size).Offset(offset)

	domains := make([]tezosdomain.TezosDomain, 0)
	if err := storage.db.GetAllByQuery(builder, &domains); err != nil {
		return tezosdomain.DomainsResponse{}, nil
	}

	countBuilder := core.NewBuilder().Count(models.DocTezosDomains).And(core.NewEq("network", network))
	doc, err := storage.db.QueryDocument(countBuilder.String())
	if err != nil {
		return tezosdomain.DomainsResponse{}, nil
	}
	var total int64
	if err := document.Scan(doc, &total); err != nil {
		return tezosdomain.DomainsResponse{}, nil
	}

	return tezosdomain.DomainsResponse{
		Domains: domains,
		Total:   total,
	}, nil
}

// ResolveDomainByAddress -
func (storage *Storage) ResolveDomainByAddress(network string, address string) (*tezosdomain.TezosDomain, error) {
	if !helpers.IsAddress(address) {
		return nil, errors.Errorf("Invalid address: %s", address)
	}

	builder := core.NewBuilder().SelectAll(models.DocTezosDomains).And(
		core.NewEq("network", network),
		core.NewEq("address", address),
	)
	var td tezosdomain.TezosDomain
	err := storage.db.GetOne(builder, &td)
	return &td, err
}
