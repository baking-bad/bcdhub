package bulk

import (
	"strings"

	"github.com/baking-bad/bcdhub/internal/genji/core"
	"github.com/baking-bad/bcdhub/internal/models"
	"github.com/baking-bad/bcdhub/internal/models/contract"
)

// Storage -
type Storage struct {
	db *core.Genji
}

// NewStorage -
func NewStorage(db *core.Genji) *Storage {
	return &Storage{db}
}

// Insert -
func (storage *Storage) Insert(items []models.Model) error {
	if len(items) == 0 {
		return nil
	}
	var bulk strings.Builder
	for i := range items {
		bulk.WriteString("INSERT INTO ")
		bulk.WriteString(items[i].GetIndex())
		bulk.WriteString(" VALUES ?")
		bulk.WriteByte(';')
		if (i%1000 == 0 && i > 0) || i == len(items)-1 {
			if err := storage.db.Exec(bulk.String()); err != nil {
				return err
			}
			bulk.Reset()
		}
	}
	return nil
}

// Update -
func (storage *Storage) Update(updates []models.Model) error {
	if len(updates) == 0 {
		return nil
	}
	return nil
}

// Delete -
func (storage *Storage) Delete(updates []models.Model) error {
	if len(updates) == 0 {
		return nil
	}
	builder := core.NewBuilder()
	for i := range updates {
		builder.Delete(updates[i].GetIndex()).And(
			core.NewEq("id", updates[i].GetID()),
		).End()
		if (i%1000 == 0 && i > 0) || i == len(updates)-1 {
			if err := storage.db.Exec(builder.String()); err != nil {
				return err
			}
			builder = core.NewBuilder()
		}
	}
	return nil
}

// RemoveField -
func (storage *Storage) RemoveField(script string, where []models.Model) error {
	if len(where) == 0 {
		return nil
	}
	return nil
}

// UpdateField -
func (storage *Storage) UpdateField(where []contract.Contract, fields ...string) error {
	if len(where) == 0 {
		return nil
	}
	return nil
}
