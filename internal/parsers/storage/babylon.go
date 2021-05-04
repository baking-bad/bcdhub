package storage

import (
	stdJSON "encoding/json"

	"github.com/baking-bad/bcdhub/internal/bcd/ast"
	"github.com/baking-bad/bcdhub/internal/models"
	"github.com/baking-bad/bcdhub/internal/models/bigmapaction"
	"github.com/baking-bad/bcdhub/internal/models/bigmapdiff"
	"github.com/baking-bad/bcdhub/internal/models/operation"
	"github.com/baking-bad/bcdhub/internal/models/types"
	"github.com/baking-bad/bcdhub/internal/noderpc"
	"github.com/pkg/errors"
)

// Babylon -
type Babylon struct {
	repo bigmapdiff.Repository
	rpc  noderpc.INode

	ptrMap            map[int64]int64
	temporaryPointers map[int64]*ast.BigMap
}

// NewBabylon -
func NewBabylon(repo bigmapdiff.Repository, rpc noderpc.INode) *Babylon {
	return &Babylon{
		repo: repo,
		rpc:  rpc,

		ptrMap:            make(map[int64]int64),
		temporaryPointers: make(map[int64]*ast.BigMap),
	}
}

// ParseTransaction -
func (b *Babylon) ParseTransaction(content noderpc.Operation, operation operation.Operation) (RichStorage, error) {
	result := content.GetResult()
	if result == nil {
		return RichStorage{Empty: true}, nil
	}

	modelUpdates, err := b.handleBigMapDiff(result, *content.Destination, operation, result.Storage)
	if err != nil {
		return RichStorage{Empty: true}, err
	}

	return RichStorage{
		Models:          modelUpdates,
		DeffatedStorage: result.Storage,
	}, nil
}

// ParseOrigination -
func (b *Babylon) ParseOrigination(content noderpc.Operation, operation operation.Operation) (RichStorage, error) {
	result := content.GetResult()
	if result == nil {
		return RichStorage{Empty: true}, nil
	}

	var scriptData struct {
		Storage stdJSON.RawMessage `json:"storage"`
	}
	if err := json.Unmarshal(content.Script, &scriptData); err != nil {
		return RichStorage{Empty: true}, err
	}

	modelUpdates, err := b.handleBigMapDiff(result, result.Originated[0], operation, scriptData.Storage)
	if err != nil {
		return RichStorage{Empty: true}, err
	}

	return RichStorage{
		Models:          modelUpdates,
		DeffatedStorage: scriptData.Storage,
	}, nil
}

func (b *Babylon) initPointersTypes(result *noderpc.OperationResult, operation operation.Operation, data []byte) error {
	level := operation.Level
	if operation.IsTransaction() && operation.Level >= 2 {
		level = operation.Level - 1
	}

	storage, err := operation.AST.StorageType()
	if err != nil {
		return err
	}

	if err := storage.SettleFromBytes(data); err != nil {
		return errors.Wrapf(err, "settleStorage %s %d", operation.Destination, level)
	}

	if err := b.checkPointers(result, storage); err == nil {
		return nil
	}

	rawData, err := b.rpc.GetScriptStorageRaw(operation.Destination, level)
	if err != nil {
		return errors.Wrapf(err, "GetScriptStorageRaw %s %d", operation.Destination, level)
	}

	if err := storage.SettleFromBytes(rawData); err != nil {
		return errors.Wrapf(err, "Settle %s %d", operation.Destination, level)
	}

	return b.checkPointers(result, storage)
}

func (b *Babylon) checkPointers(result *noderpc.OperationResult, storage *ast.TypedAst) error {
	types := storage.FindBigMapByPtr()
	for _, bmd := range result.BigMapDiffs {
		if bmd.BigMap != nil {
			ptr := *bmd.BigMap
			if typ, ok := types[ptr]; ok {
				b.temporaryPointers[ptr] = typ
				continue
			}
			if ptr < 0 {
				continue
			}
		}
		if bmd.SourceBigMap != nil {
			ptr := *bmd.SourceBigMap
			if typ, ok := types[ptr]; ok {
				b.temporaryPointers[ptr] = typ
				continue
			}
			if ptr < 0 {
				continue
			}
		}

		return ErrUnknownTemporaryPointer
	}

	return nil
}

func (b *Babylon) handleBigMapDiff(result *noderpc.OperationResult, address string, op operation.Operation, storageData []byte) ([]models.Model, error) {
	if len(result.BigMapDiffs) == 0 {
		return []models.Model{}, nil
	}

	if err := b.initPointersTypes(result, op, storageData); err != nil {
		return []models.Model{}, nil
	}

	storageModels := make([]models.Model, 0)

	handlers := map[string]func(noderpc.BigMapDiff, string, operation.Operation) ([]models.Model, error){
		"update": b.handleBigMapDiffUpdate,
		"copy":   b.handleBigMapDiffCopy,
		"remove": b.handleBigMapDiffRemove,
		"alloc":  b.handleBigMapDiffAlloc,
	}

	for i := range result.BigMapDiffs {
		action := result.BigMapDiffs[i].Action
		handler, ok := handlers[action]
		if !ok {
			continue
		}
		data, err := handler(result.BigMapDiffs[i], address, op)
		if err != nil {
			return nil, err
		}
		if len(data) > 0 {
			storageModels = append(storageModels, data...)
		}
	}
	return storageModels, nil
}

func (b *Babylon) handleBigMapDiffUpdate(item noderpc.BigMapDiff, address string, operation operation.Operation) ([]models.Model, error) {
	ptr := *item.BigMap

	bmd := bigmapdiff.BigMapDiff{
		Ptr:              ptr,
		Key:              types.Bytes(item.Key),
		KeyHash:          item.KeyHash,
		OperationHash:    operation.Hash,
		OperationCounter: operation.Counter,
		OperationNonce:   operation.Nonce,
		Level:            operation.Level,
		Contract:         address,
		Network:          operation.Network,
		Timestamp:        operation.Timestamp,
		Protocol:         operation.Protocol,
		Value:            types.Bytes(item.Value),
	}

	if err := b.addDiff(&bmd, ptr); err != nil {
		return nil, err
	}
	if ptr > -1 {
		return []models.Model{&bmd, bmd.ToState()}, nil
	}
	return nil, nil
}

func (b *Babylon) handleBigMapDiffCopy(item noderpc.BigMapDiff, address string, operation operation.Operation) ([]models.Model, error) {
	sourcePtr := *item.SourceBigMap
	destinationPtr := *item.DestBigMap

	newUpdates := make([]models.Model, 0)

	if destinationPtr > -1 {
		var srcPtr int64
		if sourcePtr > -1 {
			srcPtr = sourcePtr
		} else {
			bufPtr, err := b.getSourcePtr(sourcePtr)
			if err != nil {
				return nil, err
			}
			srcPtr = bufPtr
		}
		newUpdates = append(newUpdates, b.createBigMapDiffAction("copy", address, &srcPtr, &destinationPtr, operation))
	}

	b.ptrMap[destinationPtr] = sourcePtr

	bmd, err := b.getCopyBigMapDiff(sourcePtr, address, operation.Network)
	if err != nil {
		return nil, err
	}

	b.updateTemporaryPointers(sourcePtr, destinationPtr)

	if len(bmd) > 0 {
		for i := range bmd {
			bmd[i].Ptr = destinationPtr
			bmd[i].Contract = address
			bmd[i].Level = operation.Level
			bmd[i].Timestamp = operation.Timestamp
			bmd[i].OperationHash = operation.Hash
			bmd[i].OperationCounter = operation.Counter
			bmd[i].OperationNonce = operation.Nonce

			if err := b.addDiff(&bmd[i], destinationPtr); err != nil {
				return nil, err
			}

			if destinationPtr > -1 {
				newUpdates = append(newUpdates, &bmd[i], bmd[i].ToState())
			}
		}
	}
	return newUpdates, nil
}

func (b *Babylon) handleBigMapDiffRemove(item noderpc.BigMapDiff, address string, operation operation.Operation) ([]models.Model, error) {
	ptr := *item.BigMap
	if ptr < 0 {
		return nil, nil
	}
	states, err := b.repo.GetByPtr(operation.Network, address, ptr)
	if err != nil {
		return nil, err
	}
	newUpdates := make([]models.Model, 0)
	for i := range states {
		states[i].Removed = true

		bmd := states[i].ToDiff()
		bmd.OperationHash = operation.Hash
		bmd.OperationCounter = operation.Counter
		bmd.OperationNonce = operation.Nonce
		bmd.Level = operation.Level
		bmd.Timestamp = operation.Timestamp
		newUpdates = append(newUpdates, &bmd, &states[i])
	}
	newUpdates = append(newUpdates, b.createBigMapDiffAction("remove", address, &ptr, nil, operation))
	return newUpdates, nil
}

func (b *Babylon) handleBigMapDiffAlloc(item noderpc.BigMapDiff, address string, operation operation.Operation) ([]models.Model, error) {
	ptr := *item.BigMap

	bigMap, err := createBigMapAst(item.KeyType, item.ValueType, ptr)
	if err != nil {
		return nil, err
	}

	b.temporaryPointers[ptr] = bigMap

	var models []models.Model
	if ptr > -1 {
		models = append(
			models,
			b.createBigMapDiffAction("alloc", address, &ptr, nil, operation),
		)
	}

	return models, nil
}

func (b *Babylon) getDiffsFromUpdates(ptr int64) ([]bigmapdiff.BigMapDiff, error) {
	bigMap, ok := b.temporaryPointers[ptr]
	if !ok {
		return nil, errors.Wrapf(ErrUnknownTemporaryPointer, "%d", ptr)
	}
	diffs := bigMap.GetDiffs()
	return getBigMapDiffModels(diffs), nil
}

func (b *Babylon) createBigMapDiffAction(action, address string, srcPtr, dstPtr *int64, operation operation.Operation) *bigmapaction.BigMapAction {
	entity := &bigmapaction.BigMapAction{
		Action:      action,
		OperationID: operation.ID,
		Level:       operation.Level,
		Address:     address,
		Network:     operation.Network,
		Timestamp:   operation.Timestamp,
	}

	if srcPtr != nil && *srcPtr > -1 {
		entity.SourcePtr = srcPtr
	}

	if dstPtr != nil && *dstPtr > -1 {
		entity.DestinationPtr = dstPtr
	}

	return entity
}

func (b *Babylon) addDiff(bmd *bigmapdiff.BigMapDiff, ptr int64) error {
	if bm, ok := b.temporaryPointers[ptr]; ok {
		return bm.EnrichBigMap(prepareBigMapDiffsToEnrich([]bigmapdiff.BigMapDiff{*bmd}, false))
	}

	bigMap, ok := b.temporaryPointers[ptr]
	if !ok {
		return errors.Errorf("Can't find big map pointer: %d", ptr)
	}

	diffs := prepareBigMapDiffsToEnrich([]bigmapdiff.BigMapDiff{*bmd}, false)
	bigMap.AddDiffs(diffs...)
	b.temporaryPointers[bmd.Ptr] = bigMap
	return nil
}

func (b *Babylon) getSourcePtr(ptr int64) (int64, error) {
	if src, ok := b.ptrMap[ptr]; ok {
		return src, nil
	}
	if _, ok := b.temporaryPointers[ptr]; ok {
		return ptr, nil
	}
	return ptr, errors.Wrapf(ErrUnknownTemporaryPointer, "%d", ptr)
}

func (b *Babylon) updateTemporaryPointers(src, dst int64) {
	bigMap, ok := b.temporaryPointers[src]
	if !ok {
		return
	}
	dstBigMap := ast.Copy(bigMap).(*ast.BigMap)
	dstBigMap.Ptr = &dst
	b.temporaryPointers[dst] = dstBigMap
}

func (b *Babylon) getCopyBigMapDiff(src int64, address, network string) (bmd []bigmapdiff.BigMapDiff, err error) {
	if src > -1 {
		states, err := b.repo.GetByPtr(network, address, src)
		if err != nil {
			return nil, err
		}
		bmd = make([]bigmapdiff.BigMapDiff, 0, len(states))
		for i := range states {
			bmd = append(bmd, states[i].ToDiff())
		}
	} else {
		bmd, err = b.getDiffsFromUpdates(src)
		if err != nil {
			return nil, err
		}
	}
	return
}
