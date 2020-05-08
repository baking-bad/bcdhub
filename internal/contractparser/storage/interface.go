package storage

import (
	"github.com/baking-bad/bcdhub/internal/contractparser/meta"
	"github.com/baking-bad/bcdhub/internal/elastic"
	"github.com/baking-bad/bcdhub/internal/models"
	"github.com/tidwall/gjson"
)

// RichStorage -
type RichStorage struct {
	DeffatedStorage string
	BigMapDiffs     []*models.BigMapDiff
	Empty           bool
}

// GetBigMapDiffModels -
func (r RichStorage) GetBigMapDiffModels() []elastic.Model {
	result := make([]elastic.Model, len(r.BigMapDiffs))
	for i := range r.BigMapDiffs {
		result[i] = r.BigMapDiffs[i]
	}
	return result
}

// Parser -
type Parser interface {
	ParseTransaction(content gjson.Result, metadata meta.Metadata, operation models.Operation) (RichStorage, error)
	ParseOrigination(content gjson.Result, metadata meta.Metadata, operation models.Operation) (RichStorage, error)
	Enrich(string, []models.BigMapDiff, bool) (gjson.Result, error)

	SetUpdates(map[int64][]*models.BigMapDiff)
}
