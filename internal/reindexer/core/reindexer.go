package core

import (
	"fmt"
	"reflect"

	"github.com/baking-bad/bcdhub/internal/models"
	"github.com/restream/reindexer"
)

// Reindexer -
type Reindexer struct {
	*reindexer.Reindexer
}

// New -
func New(uri string) (*Reindexer, error) {
	db := reindexer.NewReindex(uri)
	return &Reindexer{db}, nil
}

// CreateIndexes -
func (r *Reindexer) CreateIndexes() error {
	for _, index := range models.AllModels() {
		if err := r.OpenNamespace(index.GetIndex(), reindexer.DefaultNamespaceOptions(), index); err != nil {
			return err
		}
	}
	return nil
}

// DeleteByLevelAndNetwork -
func (r *Reindexer) DeleteByLevelAndNetwork(indices []string, network string, maxLevel int64) error {
	for i := range indices {
		val := r.ExecSQL(fmt.Sprintf("DELETE FROM %s WHERE network = '%s' AND level > %d", indices[i], network, maxLevel))
		if val.Error() != nil {
			return val.Error()
		}
	}
	return nil
}

// DeleteIndices -
func (r *Reindexer) DeleteIndices(indices []string) error {
	for i := range indices {
		if err := r.DropNamespace(indices[i]); err != nil {
			return err
		}
	}
	return nil
}

// DeleteByContract -
func (r *Reindexer) DeleteByContract(indices []string, network, address string) error {
	for i := range indices {
		val := r.ExecSQL(fmt.Sprintf("DELETE FROM %s WHERE network = '%s' AND contract = '%s'", indices[i], network, address))
		if val.Error() != nil {
			return val.Error()
		}
	}
	return nil
}

// GetUnique -
func (r *Reindexer) GetUnique(field string, query *reindexer.Query) ([]string, error) {
	it := query.Distinct(field).Exec()
	defer it.Close()

	if it.Error() != nil {
		return nil, it.Error()
	}

	return it.AggResults()[0].Distincts, nil
}

// BulkInsert -
func (r *Reindexer) BulkInsert(items []models.Model) error {
	if len(items) == 0 {
		return nil
	}

	for i := range items {
		if _, err := r.Insert(items[i].GetIndex(), items[i]); err != nil {
			return err
		}
	}
	return nil
}

// BulkUpdate -
func (r *Reindexer) BulkUpdate(updates []models.Model) error {
	if len(updates) == 0 {
		return nil
	}
	for i := range updates {
		if _, err := r.Update(updates[i].GetIndex(), updates[i]); err != nil {
			return err
		}
	}
	return nil
}

// BulkDelete -
func (r *Reindexer) BulkDelete(updates []models.Model) error {
	if len(updates) == 0 {
		return nil
	}
	for i := range updates {
		if err := r.Delete(updates[i].GetIndex(), updates[i]); err != nil {
			return err
		}
	}
	return nil
}

// BulkRemoveField -
func (r *Reindexer) BulkRemoveField(field string, where []models.Model) error {
	if len(where) == 0 {
		return nil
	}
	for i := range where {
		it := r.Query(where[i].GetIndex()).Match("id", where[i].GetID()).Drop(field).Update()
		defer it.Close()

		if it.Error() != nil {
			return it.Error()
		}
	}
	return nil
}

// GetFieldValue -
func (r *Reindexer) GetFieldValue(data interface{}, field string) interface{} {
	val := reflect.ValueOf(data)
	f := reflect.Indirect(val).FieldByName(field)
	return f.Interface()
}

// SetAlias -
func (r *Reindexer) SetAlias(address, alias string) error {
	return nil
}
