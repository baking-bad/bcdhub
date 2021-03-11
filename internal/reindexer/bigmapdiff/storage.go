package bigmapdiff

import (
	"fmt"

	"github.com/baking-bad/bcdhub/internal/models"
	"github.com/baking-bad/bcdhub/internal/models/bigmapdiff"
	"github.com/baking-bad/bcdhub/internal/reindexer/core"
	"github.com/pkg/errors"
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

// CurrentByKey -
func (storage *Storage) CurrentByKey(network, keyHash string, ptr int64) (data bigmapdiff.BigMapDiff, err error) {
	if ptr < 0 {
		err = errors.Errorf("Invalid pointer value: %d", ptr)
		return
	}

	query := storage.db.Query(models.DocBigMapDiff).
		Match("network", network).
		Match("key_hash", keyHash).
		WhereInt64("ptr", reindexer.EQ, ptr).
		Sort("level", true)

	err = storage.db.GetOne(query, &data)
	return
}

// GetForAddress -
func (storage *Storage) GetForAddress(address string) ([]bigmapdiff.BigMapDiff, error) {
	query := storage.db.Query(models.DocBigMapDiff).
		Match("address", address).
		Sort("indexed_time", true)

	return storage.getTop(query, func(bmd bigmapdiff.BigMapDiff) string {
		return bmd.KeyHash
	})
}

// GetByAddress -
func (storage *Storage) GetByAddress(network, address string) (response []bigmapdiff.BigMapDiff, err error) {
	query := storage.db.Query(models.DocBigMapDiff).
		Match("network", network).
		Match("address", address).
		Sort("indexed_time", true)

	err = storage.db.GetAllByQuery(query, &response)
	return
}

// GetValuesByKey -
func (storage *Storage) GetValuesByKey(keyHash string) ([]bigmapdiff.BigMapDiff, error) {
	query := storage.db.Query(models.DocBigMapDiff).
		Match("key_hash", keyHash).
		Sort("indexed_time", true)

	return storage.getTop(query, func(bmd bigmapdiff.BigMapDiff) string {
		return fmt.Sprintf("%s_%s_%d", bmd.Network, bmd.Address, bmd.Ptr)
	})
}

// Count -
func (storage *Storage) Count(network string, ptr int64) (int64, error) {
	query := storage.db.Query(models.DocBigMapDiff).
		Distinct("key_hash").
		Match("network", network).
		WhereInt64("ptr", reindexer.EQ, ptr)

	return storage.db.Count(query)
}

// Previous -
func (storage *Storage) Previous(filters []bigmapdiff.BigMapDiff, indexedTime int64, address string) ([]bigmapdiff.BigMapDiff, error) {
	query := storage.db.Query(models.DocBigMapDiff).
		Match("address", address).
		WhereInt64("indexed_time", reindexer.LT, indexedTime)

	if len(filters) > 0 {
		query = query.OpenBracket()
		for i := range filters {
			query = query.OpenBracket().
				Match("key_hash", filters[i].KeyHash).
				CloseBracket()
			if len(filters)-1 > i {
				query = query.Or()
			}
		}
		query = query.CloseBracket()
	}
	query = query.Sort("indexed_time", true)

	return storage.getTop(query, func(bmd bigmapdiff.BigMapDiff) string {
		return fmt.Sprintf("%s_%d", bmd.KeyHash, bmd.Ptr)
	})
}

// GetUniqueByOperationID -
func (storage *Storage) GetUniqueByOperationID(operationID string) ([]bigmapdiff.BigMapDiff, error) {
	query := storage.db.Query(models.DocBigMapDiff).
		Match("operation_id", operationID).
		Sort("indexed_time", true)

	return storage.getTop(query, func(bmd bigmapdiff.BigMapDiff) string {
		return fmt.Sprintf("%d_%s", bmd.Ptr, bmd.KeyHash)
	})
}

// GetByPtrAndKeyHash -
func (storage *Storage) GetByPtrAndKeyHash(ptr int64, network, keyHash string, size, offset int64) ([]bigmapdiff.BigMapDiff, int64, error) {
	if ptr < 0 {
		return nil, 0, errors.Errorf("Invalid pointer value: %d", ptr)
	}
	if size == 0 {
		size = core.DefaultSize
	}

	query := storage.db.Query(models.DocBigMapDiff).
		Match("network", network).
		Match("key_hash", keyHash).
		WhereInt64("ptr", reindexer.EQ, ptr).
		Limit(int(size)).
		Offset(int(offset)).
		Sort("level", true)

	var total int
	result := make([]bigmapdiff.BigMapDiff, 0)
	total, err := storage.db.GetAllByQueryWithTotal(query, &result)

	return result, int64(total), err
}

// GetByOperationID -
func (storage *Storage) GetByOperationID(operationID string) ([]*bigmapdiff.BigMapDiff, error) {
	query := storage.db.Query(models.DocBigMapDiff).
		Match("operation_id", operationID)

	result := make([]*bigmapdiff.BigMapDiff, 0)
	err := storage.db.GetAllByQuery(query, &result)
	return result, err
}

// GetByPtr -
// TODO: check
func (storage *Storage) GetByPtr(address, network string, ptr int64) ([]bigmapdiff.BigMapDiff, error) {
	query := storage.db.Query(models.DocBigMapDiff).
		Match("network", network).
		Match("address", address).
		WhereInt64("ptr", reindexer.EQ, ptr)

	keyHash, err := storage.db.GetUnique("key_hash", query)
	if err != nil {
		return nil, err
	}

	secondQuery := storage.db.Query(models.DocBigMapDiff).
		Match("key_hash", keyHash...).
		Match("network", network).
		Match("address", address).
		WhereInt64("ptr", reindexer.EQ, ptr).
		Sort("indexed_time", true)

	response := make([]bigmapdiff.BigMapDiff, 0)
	err = storage.db.GetAllByQuery(secondQuery, &response)
	return response, err
}

// Get -
func (storage *Storage) Get(ctx bigmapdiff.GetContext) (response []bigmapdiff.Bucket, err error) {
	if *ctx.Ptr < 0 {
		return nil, errors.Errorf("Invalid pointer value: %d", *ctx.Ptr)
	}
	query := storage.db.Query(models.DocBigMapDiff)
	buildGetContext(ctx, query)
	err = storage.db.GetAllByQuery(query, &response)
	return
}

func (storage *Storage) getTop(query *reindexer.Query, idFunc func(bigmapdiff.BigMapDiff) string) ([]bigmapdiff.BigMapDiff, error) {
	all := make([]bigmapdiff.BigMapDiff, 0)
	if err := storage.db.GetAllByQuery(query, &all); err != nil {
		return nil, err
	}

	response := make([]bigmapdiff.BigMapDiff, 0)
	found := make(map[string]struct{})
	for i := range all {
		id := idFunc(all[i])
		if _, ok := found[id]; ok {
			continue
		}
		found[id] = struct{}{}
		response = append(response, all[i])
	}
	return response, nil
}
