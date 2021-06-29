package storage

import (
	"github.com/baking-bad/bcdhub/internal/bcd"
	"github.com/baking-bad/bcdhub/internal/bcd/ast"
	"github.com/baking-bad/bcdhub/internal/models"
	"github.com/baking-bad/bcdhub/internal/models/bigmap"
	"github.com/baking-bad/bcdhub/internal/noderpc"
	"github.com/pkg/errors"
)

// Errors -
var (
	ErrInvalidPath          = errors.Errorf("Invalid path")
	ErrPathIsNotPointer     = errors.Errorf("Path is not pointer")
	ErrPointerAlreadyExists = errors.Errorf("Pointer already exists")
)

// Enrich -
func Enrich(storage *ast.TypedAst, bmd []bigmap.Diff, skipEmpty, unpack bool) error {
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
func EnrichFromState(storage *ast.TypedAst, bmd []bigmap.State, skipEmpty, unpack bool) error {
	if len(bmd) == 0 {
		return nil
	}
	if !storage.IsSettled() {
		return ErrTreeIsNotSettled
	}

	data := prepareBigMapStatesToEnrich(bmd, skipEmpty)
	return storage.EnrichBigMap(data)
}

// MakeStorageParser -
func MakeStorageParser(bigmaps bigmap.Repository, stateRepo bigmap.StateRepository, general models.GeneralRepository, rpc noderpc.INode, protocol string) (Parser, error) {
	protoSymLink, err := bcd.GetProtoSymLink(protocol)
	if err != nil {
		return nil, err
	}

	switch protoSymLink {
	case bcd.SymLinkBabylon:
		return NewBabylon(bigmaps, stateRepo, general, rpc), nil
	case bcd.SymLinkAlpha:
		return NewAlpha(), nil
	default:
		return nil, errors.Errorf("Unknown protocol %s", protocol)
	}
}
