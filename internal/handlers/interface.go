package handlers

import (
	"github.com/baking-bad/bcdhub/internal/bcd/ast"
	"github.com/baking-bad/bcdhub/internal/models"
	"github.com/baking-bad/bcdhub/internal/models/domains"
)

// Handler -
type Handler[M models.Constraint] interface {
	Do(bmd *domains.BigMapDiff, storage *ast.TypedAst) ([]M, error)
}
