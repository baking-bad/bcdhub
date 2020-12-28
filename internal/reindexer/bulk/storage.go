package bulk

import (
	"reflect"

	"github.com/baking-bad/bcdhub/internal/models"
	"github.com/baking-bad/bcdhub/internal/models/contract"
	"github.com/baking-bad/bcdhub/internal/reindexer/core"
)

// Storage -
type Storage struct {
	db *core.Reindexer
}

// NewStorage -
func NewStorage(db *core.Reindexer) *Storage {
	return &Storage{db}
}

// Insert -
func (storage *Storage) Insert(items []models.Model) error {
	if len(items) == 0 {
		return nil
	}

	for i := range items {
		if _, err := storage.db.Insert(items[i].GetIndex(), items[i]); err != nil {
			return err
		}
	}
	return nil
}

// Update -
func (storage *Storage) Update(updates []models.Model) error {
	if len(updates) == 0 {
		return nil
	}
	for i := range updates {
		if _, err := storage.db.Update(updates[i].GetIndex(), updates[i]); err != nil {
			return err
		}
	}
	return nil
}

// Delete -
func (storage *Storage) Delete(updates []models.Model) error {
	if len(updates) == 0 {
		return nil
	}
	for i := range updates {
		if err := storage.db.Delete(updates[i].GetIndex(), updates[i]); err != nil {
			return err
		}
	}
	return nil
}

// RemoveField -
func (storage *Storage) RemoveField(field string, where []models.Model) error {
	if len(where) == 0 {
		return nil
	}
	for i := range where {
		it := storage.db.Query(where[i].GetIndex()).Match("id", where[i].GetID()).Drop(field).Update()
		defer it.Close()

		if it.Error() != nil {
			return it.Error()
		}
	}
	return nil
}

// UpdateField -
func (storage *Storage) UpdateField(where []contract.Contract, fields ...string) error {
	if len(where) == 0 {
		return nil
	}
	tx, err := storage.db.BeginTx(models.DocContracts)
	if err != nil {
		return err
	}
	for i := range where {
		query := tx.Query().Match("id", where[i].GetID())
		for j := range fields {
			value := storage.getFieldValue(where[i], fields[j])
			query = query.Set(fields[j], value)
		}
		it := query.Update()
		defer it.Close()

		if it.Error() != nil {
			return it.Error()
		}
	}
	return tx.Commit()
}

func (storage *Storage) getFieldValue(c contract.Contract, field string) interface{} {
	r := reflect.ValueOf(c)
	f := reflect.Indirect(r).FieldByName(field)
	return f.Interface()
}
