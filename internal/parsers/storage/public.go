package storage

import (
	"github.com/baking-bad/bcdhub/internal/bcd/ast"
	"github.com/baking-bad/bcdhub/internal/models/bigmapdiff"
	"github.com/pkg/errors"
)

// Errors -
var (
	ErrInvalidPath          = errors.Errorf("Invalid path")
	ErrPathIsNotPointer     = errors.Errorf("Path is not pointer")
	ErrPointerAlreadyExists = errors.Errorf("Pointer already exists")
)

// Enrich -
func Enrich(storage *ast.TypedAst, bmd []bigmapdiff.BigMapDiff, skipEmpty, unpack bool) error {
	if len(bmd) == 0 {
		return nil
	}
	if !storage.IsSettled() {
		return ErrTreeIsNotSettled
	}

	data := prepareBigMapDiffsToEnrich(bmd, skipEmpty)
	return storage.EnrichBigMap(data)
}

// EnrichFromState -
func EnrichFromState(storage *ast.TypedAst, bmd []bigmapdiff.BigMapState, skipEmpty, unpack bool) error {
	if len(bmd) == 0 {
		return nil
	}
	if !storage.IsSettled() {
		return ErrTreeIsNotSettled
	}

	data := prepareBigMapStatesToEnrich(bmd, skipEmpty)
	return storage.EnrichBigMap(data)
}
