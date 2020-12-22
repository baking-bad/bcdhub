package protocol

import (
	"fmt"

	"github.com/baking-bad/bcdhub/internal/genji/core"
	"github.com/baking-bad/bcdhub/internal/models"
	"github.com/baking-bad/bcdhub/internal/models/protocol"
	"github.com/genjidb/genji/document"
)

// Storage -
type Storage struct {
	db *core.Genji
}

// NewStorage -
func NewStorage(db *core.Genji) *Storage {
	return &Storage{db}
}

// GetProtocol - returns current protocol for `network` and `level` (`hash` is optional, leave empty string for default)
func (storage *Storage) GetProtocol(network, hash string, level int64) (p protocol.Protocol, err error) {
	builder := core.NewBuilder().SelectAll(models.DocProtocol)

	conditions := make([]fmt.Stringer, 0)
	conditions = append(conditions, core.NewEq("network", network))
	if level > -1 {
		conditions = append(conditions, core.NewLte("start_level", level))
	}
	if hash != "" {
		conditions = append(conditions, core.NewEq("hash", hash))
	}
	builder.And(conditions...).SortDesc("start_level")

	err = storage.db.GetOne(builder, &p)
	return
}

// GetSymLinks - returns list of symlinks in `network` after `level`
func (storage *Storage) GetSymLinks(network string, level int64) (map[string]struct{}, error) {
	builder := core.NewBuilder().SelectAll(models.DocProtocol).And(
		core.NewEq("network", network),
		core.NewGt("start_level", level),
	).SortDesc("start_level")

	symMap := make(map[string]struct{})

	res, err := storage.db.Query(builder.String())
	if err != nil {
		return nil, err
	}
	defer res.Close()

	res.Iterate(func(d document.Document) error {
		val, err := d.GetByField("sym_link")
		if err != nil {
			return err
		}
		symMap[val.String()] = struct{}{}
		return nil
	})

	return symMap, nil
}
