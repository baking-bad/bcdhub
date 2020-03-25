package storage

import (
	"github.com/baking-bad/bcdhub/internal/contractparser/meta"
	"github.com/baking-bad/bcdhub/internal/models"
	"github.com/tidwall/gjson"
)

// RichStorage -
type RichStorage struct {
	DeffatedStorage string
	BigMapDiffs     []models.BigMapDiff
	Empty           bool
}

// Parser -
type Parser interface {
	ParseTransaction(content gjson.Result, metadata meta.Metadata, level int64, operationID string) (RichStorage, error)
	ParseOrigination(content gjson.Result, metadata meta.Metadata, level int64, operationID string) (RichStorage, error)
	Enrich(string, []models.BigMapDiff, bool) (gjson.Result, error)
}
