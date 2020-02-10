package storage

import (
	"github.com/aopoltorzhicky/bcdhub/internal/models"
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
	ParseTransaction(content gjson.Result, protocol string, level int64, operationID string) (RichStorage, error)
	ParseOrigination(content gjson.Result, protocol string, level int64, operationID string) (RichStorage, error)
	Enrich(string, gjson.Result) (gjson.Result, error)
}
