package transfer

import (
	"fmt"

	"github.com/baking-bad/bcdhub/internal/bcd/consts"
	"github.com/baking-bad/bcdhub/internal/models"
	"github.com/baking-bad/bcdhub/internal/models/transfer"
	"github.com/baking-bad/bcdhub/internal/models/tzip"
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

const (
	maxTransfersSize = 10000
)

// Get -
func (storage *Storage) Get(ctx transfer.GetContext) (po transfer.Pageable, err error) {
	query := storage.db.Query(models.DocTransfers)
	buildGetContext(ctx, query)

	transfers := make([]transfer.Transfer, 0)
	total, err := storage.db.GetAllByQueryWithTotal(query, &transfers)
	if err != nil {
		return
	}

	po.Transfers = transfers
	po.Total = int64(total)

	if len(transfers) > 0 {
		po.LastID = fmt.Sprintf("%d", transfers[len(transfers)-1].IndexedTime)
	}
	return
}

// GetAll -
func (storage *Storage) GetAll(network string, level int64) (transfers []transfer.Transfer, err error) {
	query := storage.db.Query(models.DocTransfers).
		Match("network", network).
		WhereInt64("level", reindexer.GT, level)

	err = storage.db.GetAllByQuery(query, &transfers)
	return
}

// GetTokenSupply -
func (storage *Storage) GetTokenSupply(network, address string, tokenID int64) (result transfer.TokenSupply, err error) {
	it := storage.db.Query(models.DocTransfers).
		Match("network", network).
		Match("contract", address).
		Match("status", consts.Applied).
		WhereInt64("token_id", reindexer.EQ, tokenID).
		Exec()
	defer it.Close()

	if it.Error() != nil {
		return result, it.Error()
	}

	for it.Next() {
		var t transfer.Transfer
		it.NextObj(&t)

		switch {
		case t.From == "":
			result.Supply += t.Amount
		case t.To == "":
			result.Supply -= t.Amount
		default:
			result.Transfered += t.Amount
		}
	}
	return
}

// GetTokenVolumeSeries -
func (storage *Storage) GetTokenVolumeSeries(network, period string, contracts []string, entrypoints []tzip.DAppContract, tokenID uint) ([][]float64, error) {
	return nil, nil
}

// GetToken24HoursVolume -
func (storage *Storage) GetToken24HoursVolume(network, contract string, initiators, entrypoints []string, tokenID int64) (float64, error) {
	return 0, nil
}
