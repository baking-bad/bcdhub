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
	storage, err := getStorage(operation)
	if err != nil {
		return RichStorage{Empty: true}, err
	}
	result := content.GetResult()
	if result == nil {
		return RichStorage{Empty: true}, nil
	}

	level := operation.Level - 1
	if operation.Level < 2 {
		level = operation.Level
	}

	if err := b.initPointersTypes(result, storage, operation.Destination, level, result.Storage); err != nil {
		return RichStorage{Empty: true}, nil
	}

	modelUpdates, err := b.handleBigMapDiff(result, *content.Destination, operation)
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
	storage, err := getStorage(operation)
	if err != nil {
		return RichStorage{Empty: true}, err
	}
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

	if err := b.initPointersTypes(result, storage, operation.Destination, operation.Level, scriptData.Storage); err != nil {
		return RichStorage{Empty: true}, nil
	}

	modelUpdates, err := b.handleBigMapDiff(result, result.Originated[0], operation)
	if err != nil {
		return RichStorage{Empty: true}, err
	}

	return RichStorage{
		Models:          modelUpdates,
		DeffatedStorage: scriptData.Storage,
	}, nil
}

func (b *Babylon) initPointersTypes(result *noderpc.OperationResult, storage *ast.TypedAst, address string, level int64, data []byte) error {
	var storageData ast.UntypedAST
	if err := json.Unmarshal(data, &storageData); err != nil {
		return errors.Wrapf(err, "settleStorage %s %d", address, level)
	}

	if err := storage.Settle(storageData); err != nil {
		return errors.Wrapf(err, "settleStorage %s %d", address, level)
	}

	if err := b.checkPointers(result, storage); err == nil {
		return nil
	}

	rawData, err := b.rpc.GetScriptStorageRaw(address, level)
	if err != nil {
		return errors.Wrapf(err, "settleStorage %s %d", address, level)
	}

	var nodeStorageData ast.UntypedAST
	if err := json.Unmarshal(rawData, &nodeStorageData); err != nil {
		return errors.Wrapf(err, "settleStorage %s %d", address, level)
	}

	if err := storage.Settle(nodeStorageData); err != nil {
		return errors.Wrapf(err, "settleStorage %s %d", address, level)
	}

	if err := b.checkPointers(result, storage); err != nil {
		return errors.Wrapf(err, "settleStorage %s %d", address, level)
	}

	return nil
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

func (b *Babylon) handleBigMapDiff(result *noderpc.OperationResult, address string, op operation.Operation) ([]models.Model, error) {
	if len(result.BigMapDiffs) == 0 {
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
		return []models.Model{&bmd}, nil
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
				newUpdates = append(newUpdates, &bmd[i])
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
	bmd, err := b.repo.GetByPtr(address, operation.Network, ptr)
	if err != nil {
		return nil, err
	}
	newUpdates := make([]models.Model, len(bmd))
	for i := range bmd {
		bmd[i].OperationHash = operation.Hash
		bmd[i].OperationCounter = operation.Counter
		bmd[i].OperationNonce = operation.Nonce
		bmd[i].Level = operation.Level
		bmd[i].Timestamp = operation.Timestamp
		bmd[i].Value = nil
		newUpdates[i] = &bmd[i]
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
		bmd, err = b.repo.GetByPtr(address, network, src)
		if err != nil {
			return nil, err
		}
	} else {
		bmd, err = b.getDiffsFromUpdates(src)
		if err != nil {
			return nil, err
		}
	}
	return
}
