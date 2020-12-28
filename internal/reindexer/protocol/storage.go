package protocol

import (
	"github.com/baking-bad/bcdhub/internal/models"
	"github.com/baking-bad/bcdhub/internal/models/protocol"
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

// GetProtocol - returns current protocol for `network` and `level` (`hash` is optional, leave empty string for default)
func (storage *Storage) GetProtocol(network, hash string, level int64) (p protocol.Protocol, err error) {
	query := storage.db.Query(models.DocProtocol).
		Match("network", network)

	if level > -1 {
		query = query.WhereInt64("start_level", reindexer.LE, level)
	}
	if hash != "" {
		query = query.Match("hash", hash)
	}
	query = query.Sort("start_level", true)

	err = storage.db.GetOne(query, &p)
	return
}

// GetSymLinks - returns list of symlinks in `network` after `level`
func (storage *Storage) GetSymLinks(network string, level int64) (map[string]struct{}, error) {
	it := storage.db.Query(models.DocProtocol).
		Select("sym_link").
		Match("network", network).
		WhereInt64("start_level", reindexer.GT, level).
		Sort("start_level", true).Exec()
	defer it.Close()

	if it.Error() != nil {
		return nil, it.Error()
	}

	symMap := make(map[string]struct{})

	type link struct {
		symLimk string `reindex:"sym_link"`
	}
	for it.Next() {
		var sl link
		it.NextObj(&sl)
		symMap[sl.symLimk] = struct{}{}
	}

	return symMap, nil
}
