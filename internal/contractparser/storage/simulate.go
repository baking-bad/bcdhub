package storage

import (
	"github.com/baking-bad/bcdhub/internal/contractparser/meta"
	"github.com/baking-bad/bcdhub/internal/elastic"
	"github.com/baking-bad/bcdhub/internal/models"
	"github.com/baking-bad/bcdhub/internal/noderpc"
	"github.com/tidwall/gjson"
)

// Simulate -
type Simulate struct {
	*Babylon
}

// NewSimulate -
func NewSimulate(rpc noderpc.INode, es elastic.IElastic) *Simulate {
	return &Simulate{
		Babylon: NewBabylon(rpc, es),
	}
}

// ParseTransaction -
func (b *Simulate) ParseTransaction(content gjson.Result, metadata meta.Metadata, operation models.Operation) (RichStorage, error) {
	storage := content.Get("storage")
	var bm []elastic.Model
	if content.Get("big_map_diff.#").Int() > 0 {
		ptrMap, err := FindBigMapPointers(metadata, storage)
		if err != nil {
			return RichStorage{Empty: true}, err
		}

		if bm, err = b.handleBigMapDiff(content, ptrMap, operation.Destination, operation); err != nil {
			return RichStorage{Empty: true}, err
		}
	}
	return RichStorage{
		Models:          bm,
		DeffatedStorage: storage.Raw,
	}, nil
}

// ParseOrigination -
func (b *Simulate) ParseOrigination(content gjson.Result, metadata meta.Metadata, operation models.Operation) (RichStorage, error) {
	storage := operation.Script.Get("storage")

	var bm []elastic.Model
	if content.Get("big_map_diff.#").Int() > 0 {
		ptrMap, err := FindBigMapPointers(metadata, storage)
		if err != nil {
			return RichStorage{Empty: true}, err
		}

		if bm, err = b.handleBigMapDiff(content, ptrMap, operation.Source, operation); err != nil {
			return RichStorage{Empty: true}, err
		}
	}

	return RichStorage{
		Models:          bm,
		DeffatedStorage: storage.String(),
	}, nil
}
