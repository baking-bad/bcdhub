package core

import (
	"fmt"

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
