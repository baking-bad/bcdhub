package search

import (
	"context"

	"github.com/baking-bad/bcdhub/internal/models"
	"github.com/baking-bad/bcdhub/internal/models/types"
)

// Searcher -
type Searcher interface {
	ByText(text string, offset int64, fields []string, filters map[string]interface{}, group bool) (Result, error)
	Save(ctx context.Context, items []Data) error
	CreateIndexes() error
	Rollback(network string, level int64) error
	BigMapDiffs(args BigMapDiffSearchArgs) ([]BigMapDiffResult, error)
}

// Data -
type Data interface {
	GetID() string
	GetIndex() string
	Prepare(network types.Network, model models.Model)
}
