package storage

import (
	"github.com/baking-bad/bcdhub/internal/contractparser/meta"
	"github.com/baking-bad/bcdhub/internal/models"
	"github.com/baking-bad/bcdhub/internal/models/bigmapdiff"
	"github.com/baking-bad/bcdhub/internal/models/operation"
	"github.com/tidwall/gjson"
)

// RichStorage -
type RichStorage struct {
	DeffatedStorage string
	Models          []models.Model
	Empty           bool
}

// Parser -
type Parser interface {
	ParseTransaction(content gjson.Result, metadata meta.Metadata, operation operation.Operation) (RichStorage, error)
	ParseOrigination(content gjson.Result, metadata meta.Metadata, operation operation.Operation) (RichStorage, error)
	Enrich(string, string, []bigmapdiff.BigMapDiff, bool, bool) (gjson.Result, error)
}
