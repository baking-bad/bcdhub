package balanceupdate

import (
	"github.com/baking-bad/bcdhub/internal/models"
	"github.com/baking-bad/bcdhub/internal/reindexer/core"
	"github.com/restream/reindexer"
)

// Storage -
type Storage struct {
	db *core.Reindexer
}

// NewStorage -
func NewStorage(db *core.Reindexer) *Storage {
	return &Storage{db}
}

// GetBalance -
func (storage *Storage) GetBalance(network, address string) (int64, error) {
	query := storage.db.Query(models.DocBalanceUpdates).
		WhereString("network", reindexer.EQ, network).
		WhereString("contract", reindexer.EQ, address)
	query.AggregateSum("change")

	it := query.Exec()
	if it.Error() != nil {
		return 0, it.Error()
	}
	agg := it.AggResults()[0]
	return int64(agg.Value), nil
}
