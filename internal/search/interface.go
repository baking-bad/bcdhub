package search

import "github.com/baking-bad/bcdhub/internal/models"

// Searcher -
type Searcher interface {
	ByText(text string, offset int64, fields []string, filters map[string]interface{}, group bool) (Result, error)
	Save(items []Data) error
	CreateIndexes() error
	Rollback(network string, level int64) error
}

// Data -
type Data interface {
	GetID() string
	GetIndex() string
	Prepare(model models.Model)
}
