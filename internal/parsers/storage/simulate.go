package storage

import (
	"github.com/baking-bad/bcdhub/internal/bcd/ast"
	"github.com/baking-bad/bcdhub/internal/models"
	"github.com/baking-bad/bcdhub/internal/models/bigmapdiff"
	"github.com/baking-bad/bcdhub/internal/models/operation"
	"github.com/tidwall/gjson"
)

// Simulate -
type Simulate struct {
	*Babylon
}

// NewSimulate -
func NewSimulate(repo bigmapdiff.Repository) *Simulate {
	return &Simulate{
		Babylon: NewBabylon(repo),
	}
}

// ParseTransaction -
func (b *Simulate) ParseTransaction(content gjson.Result, operation operation.Operation) (RichStorage, error) {
	storage, err := getStorage(operation.Script)
	if err != nil {
		return RichStorage{Empty: true}, err
	}
	storageData := content.Get("storage")
	var data ast.UntypedAST
	if err := json.UnmarshalFromString(storageData.String(), &data); err != nil {
		return RichStorage{Empty: true}, err
	}

	if err := storage.Settle(data); err != nil {
		return RichStorage{Empty: true}, err
	}

	var bm []models.Model
	if content.Get("big_map_diff.#").Int() > 0 {
		var err error
		if bm, err = b.handleBigMapDiff(content, storage, operation.Destination, operation); err != nil {
			return RichStorage{Empty: true}, err
		}
	}
	return RichStorage{
		Models:          bm,
		DeffatedStorage: storageData.Raw,
	}, nil
}

// ParseOrigination -
func (b *Simulate) ParseOrigination(content gjson.Result, operation operation.Operation) (RichStorage, error) {
	storage, err := getStorage(operation.Script)
	if err != nil {
		return RichStorage{Empty: true}, err
	}
	storageData := operation.Script.Get("storage")
	var data ast.UntypedAST
	if err := json.UnmarshalFromString(storageData.String(), &data); err != nil {
		return RichStorage{Empty: true}, err
	}

	if err := storage.Settle(data); err != nil {
		return RichStorage{Empty: true}, err
	}

	var bm []models.Model
	if content.Get("big_map_diff.#").Int() > 0 {
		var err error
		if bm, err = b.handleBigMapDiff(content, storage, operation.Source, operation); err != nil {
			return RichStorage{Empty: true}, err
		}
	}

	return RichStorage{
		Models:          bm,
		DeffatedStorage: storageData.String(),
	}, nil
}
