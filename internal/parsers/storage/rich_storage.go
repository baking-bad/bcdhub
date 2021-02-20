package storage

import (
	"github.com/baking-bad/bcdhub/internal/models"
	"github.com/baking-bad/bcdhub/internal/models/operation"
	jsoniter "github.com/json-iterator/go"
	"github.com/tidwall/gjson"
)

var json = jsoniter.ConfigCompatibleWithStandardLibrary

// RichStorage -
type RichStorage struct {
	DeffatedStorage string
	Models          []models.Model
	Empty           bool
}

// Parser -
type Parser interface {
	ParseTransaction(content gjson.Result, operation operation.Operation) (RichStorage, error)
	ParseOrigination(content gjson.Result, operation operation.Operation) (RichStorage, error)
}
