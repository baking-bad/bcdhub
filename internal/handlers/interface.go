package handlers

import (
	"github.com/baking-bad/bcdhub/internal/bcd/ast"
	"github.com/baking-bad/bcdhub/internal/models"
	"github.com/baking-bad/bcdhub/internal/models/bigmapdiff"
	jsoniter "github.com/json-iterator/go"
)

var json = jsoniter.ConfigCompatibleWithStandardLibrary

// Handler -
type Handler interface {
	Do(bmd *bigmapdiff.BigMapDiff, storage *ast.TypedAst) (bool, []models.Model, error)
}
