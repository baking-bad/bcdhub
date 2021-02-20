package storage

import (
	"github.com/baking-bad/bcdhub/internal/bcd"
	"github.com/baking-bad/bcdhub/internal/bcd/ast"
	"github.com/baking-bad/bcdhub/internal/bcd/consts"
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

// MakeStorageParser -
func MakeStorageParser(repo bigmapdiff.Repository, protocol string, isSimulating bool) (Parser, error) {
	if isSimulating {
		return NewSimulate(repo), nil
	}

	protoSymLink, err := bcd.GetProtoSymLink(protocol)
	if err != nil {
		return nil, err
	}

	switch protoSymLink {
	case consts.MetadataBabylon:
		return NewBabylon(repo), nil
	case consts.MetadataAlpha:
		return NewAlpha(), nil
	default:
		return nil, errors.Errorf("Unknown protocol %s", protocol)
	}
}
