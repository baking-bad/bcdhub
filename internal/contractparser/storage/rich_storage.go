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
	Models          []elastic.Model
	Empty           bool
}

// Parser -
type Parser interface {
	ParseTransaction(content gjson.Result, metadata meta.Metadata, operation models.Operation) (RichStorage, error)
	ParseOrigination(content gjson.Result, metadata meta.Metadata, operation models.Operation) (RichStorage, error)
	Enrich(string, string, []models.BigMapDiff, bool, bool) (gjson.Result, error)
}
