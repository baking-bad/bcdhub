package storage

import (
	"github.com/baking-bad/bcdhub/internal/bcd/ast"
	"github.com/baking-bad/bcdhub/internal/models/account"
	"github.com/baking-bad/bcdhub/internal/models/bigmapaction"
	"github.com/baking-bad/bcdhub/internal/models/bigmapdiff"
	"github.com/baking-bad/bcdhub/internal/models/operation"
	"github.com/baking-bad/bcdhub/internal/models/types"
	"github.com/baking-bad/bcdhub/internal/noderpc"
	"github.com/baking-bad/bcdhub/internal/parsers"
	"github.com/pkg/errors"
)

// Babylon -
type Babylon struct {
	bigmapdiffs bigmapdiff.Repository
	operations  operation.Repository
	accounts    account.Repository

	ptrMap            map[int64]int64
	temporaryPointers map[int64]*ast.BigMap
}

// NewBabylon -
func NewBabylon(bigmapdiffs bigmapdiff.Repository, operations operation.Repository, accounts account.Repository) *Babylon {
	return &Babylon{
		bigmapdiffs: bigmapdiffs,
		operations:  operations,
		accounts:    accounts,

		ptrMap:            make(map[int64]int64),
		temporaryPointers: make(map[int64]*ast.BigMap),
	}
}

// ParseTransaction -
func (b *Babylon) ParseTransaction(content noderpc.Operation, operation *operation.Operation, store parsers.Store) error {
	result := content.GetResult()
	if result == nil {
		return nil
	}
	operation.DeffatedStorage = result.Storage

	return b.handleBigMapDiff(result, *content.Destination, operation, result.Storage, store)
}

// ParseOrigination -
func (b *Babylon) ParseOrigination(content noderpc.Operation, operation *operation.Operation, store parsers.Store) error {
	result := content.GetResult()
	if result == nil {
		return nil
	}

	return b.handleBigMapDiff(result, result.Originated[0], operation, operation.DeffatedStorage, store)
}

func (b *Babylon) initPointersTypes(result *noderpc.OperationResult, operation *operation.Operation, data []byte) error {
	level := operation.Level
	if operation.IsTransaction() && operation.Level >= 2 {
		level = operation.Level - 1
	}

	storage, err := operation.AST.StorageType()
	if err != nil {
		return err
	}

	if err := storage.SettleFromBytes(data); err != nil {
		return errors.Wrapf(err, "settleStorage %s %d", operation.Destination.Address, level)
	}

	if err := b.checkPointers(result, storage); err == nil {
		return nil
	}

	account, err := b.accounts.Get(operation.Destination.Address)
	if err != nil {
		return err
	}

	operaiton, err := b.operations.Last(map[string]interface{}{
		"destination_id": account.ID,
		"status":         types.OperationStatusApplied,
	}, 0)
	if err != nil {
		return err
	}
	if err := storage.SettleFromBytes(operaiton.DeffatedStorage); err != nil {
		return errors.Wrapf(err, "Settle %s %d", operation.Destination.Address, level)
	}

	return b.checkPointers(result, storage)
}

func (b *Babylon) checkPointers(result *noderpc.OperationResult, storage *ast.TypedAst) error {
	bigMaps := storage.FindBigMapByPtr()
	for _, bmd := range result.BigMapDiffs {
		switch {
		case bmd.Action == types.BigMapActionStringAlloc:
			ptr := *bmd.BigMap
			typ, err := createBigMapAst(bmd.KeyType, bmd.ValueType, ptr)
			if err != nil {
				return err
			}
			bigMaps[ptr] = typ
			b.temporaryPointers[ptr] = typ
			continue

		case bmd.BigMap != nil:
			ptr := *bmd.BigMap
			if typ, ok := bigMaps[ptr]; ok {
				if _, ok := b.temporaryPointers[ptr]; !ok {
					b.temporaryPointers[ptr] = typ
				}
				continue
			}
			if ptr < 0 {
				continue
			}

		case bmd.SourceBigMap != nil:
			ptr := *bmd.SourceBigMap
			if typ, ok := bigMaps[ptr]; ok {
				if _, ok := b.temporaryPointers[ptr]; !ok {
					b.temporaryPointers[ptr] = typ
				}
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

func (b *Babylon) handleBigMapDiff(result *noderpc.OperationResult, address string, op *operation.Operation, storageData []byte, store parsers.Store) error {
	if len(result.BigMapDiffs) == 0 {
		return nil
	}

	if err := b.initPointersTypes(result, op, storageData); err != nil {
		return nil
	}

	handlers := map[string]func(noderpc.BigMapDiff, string, *operation.Operation, parsers.Store) error{
		types.BigMapActionStringUpdate: b.handleBigMapDiffUpdate,
		types.BigMapActionStringCopy:   b.handleBigMapDiffCopy,
		types.BigMapActionStringRemove: b.handleBigMapDiffRemove,
		types.BigMapActionStringAlloc:  b.handleBigMapDiffAlloc,
	}

	for i := range result.BigMapDiffs {
		action := result.BigMapDiffs[i].Action
		handler, ok := handlers[action]
		if !ok {
			continue
		}

		if err := handler(result.BigMapDiffs[i], address, op, store); err != nil {
			return err
		}
	}

	op.BigMapDiffsCount = len(op.BigMapDiffs)
	return nil
}

func (b *Babylon) handleBigMapDiffUpdate(item noderpc.BigMapDiff, address string, operation *operation.Operation, store parsers.Store) error {
	ptr := *item.BigMap

	bmd := bigmapdiff.BigMapDiff{
		Ptr:         ptr,
		Key:         types.Bytes(item.Key),
		KeyHash:     item.KeyHash,
		OperationID: operation.ID,
		Level:       operation.Level,
		Contract:    address,
		Timestamp:   operation.Timestamp,
		ProtocolID:  operation.ProtocolID,
		Value:       types.Bytes(item.Value),
	}

	if err := b.addDiff(&bmd, ptr); err != nil {
		return err
	}

	if ptr > -1 {
		operation.BigMapDiffs = append(operation.BigMapDiffs, &bmd)
		store.AddBigMapStates(bmd.ToState())
	}
	return nil
}

func (b *Babylon) handleBigMapDiffCopy(item noderpc.BigMapDiff, address string, operation *operation.Operation, store parsers.Store) error {
	sourcePtr := *item.SourceBigMap
	destinationPtr := *item.DestBigMap

	if destinationPtr > -1 {
		var srcPtr int64
		if sourcePtr > -1 {
			srcPtr = sourcePtr
		} else {
			bufPtr, err := b.getSourcePtr(sourcePtr)
			if err != nil {
				return err
			}
			srcPtr = bufPtr
		}
		operation.BigMapActions = append(operation.BigMapActions, b.createBigMapDiffAction("copy", address, &srcPtr, &destinationPtr, operation))
	}

	b.ptrMap[destinationPtr] = sourcePtr

	bmd, err := b.getCopyBigMapDiff(sourcePtr, address)
	if err != nil {
		return err
	}

	b.updateTemporaryPointers(sourcePtr, destinationPtr)

	if len(bmd) > 0 {
		for i := range bmd {
			bmd[i].Ptr = destinationPtr
			bmd[i].Contract = address
			bmd[i].Level = operation.Level
			bmd[i].Timestamp = operation.Timestamp
			bmd[i].OperationID = operation.ID
			bmd[i].ProtocolID = operation.ProtocolID

			if err := b.addDiff(&bmd[i], destinationPtr); err != nil {
				return err
			}

			if destinationPtr > -1 {
				operation.BigMapDiffs = append(operation.BigMapDiffs, &bmd[i])
				store.AddBigMapStates(bmd[i].ToState())
			}
		}
	}
	return nil
}

func (b *Babylon) handleBigMapDiffRemove(item noderpc.BigMapDiff, address string, operation *operation.Operation, store parsers.Store) error {
	ptr := *item.BigMap
	if ptr < 0 {
		return nil
	}
	states, err := b.bigmapdiffs.GetByPtr(address, ptr)
	if err != nil {
		return err
	}
	for i := range states {
		states[i].Removed = true

		bmd := states[i].ToDiff()
		bmd.OperationID = operation.ID
		bmd.Level = operation.Level
		bmd.Timestamp = operation.Timestamp
		bmd.ProtocolID = operation.ProtocolID

		operation.BigMapDiffs = append(operation.BigMapDiffs, &bmd)
		store.AddBigMapStates(&states[i])
	}
	operation.BigMapActions = append(operation.BigMapActions, b.createBigMapDiffAction("remove", address, &ptr, nil, operation))
	return nil
}

func (b *Babylon) handleBigMapDiffAlloc(item noderpc.BigMapDiff, address string, operation *operation.Operation, _ parsers.Store) error {
	ptr := *item.BigMap
	if ptr > -1 {
		operation.BigMapActions = append(operation.BigMapActions, b.createBigMapDiffAction("alloc", address, &ptr, nil, operation))
	}

	return nil
}

func (b *Babylon) getDiffsFromUpdates(ptr int64) ([]bigmapdiff.BigMapDiff, error) {
	bigMap, ok := b.temporaryPointers[ptr]
	if !ok {
		return nil, errors.Wrapf(ErrUnknownTemporaryPointer, "%d", ptr)
	}
	diffs := bigMap.GetDiffs()
	return getBigMapDiffModels(diffs), nil
}

func (b *Babylon) createBigMapDiffAction(action, address string, srcPtr, dstPtr *int64, operation *operation.Operation) *bigmapaction.BigMapAction {
	entity := &bigmapaction.BigMapAction{
		Action:      types.NewBigMapAction(action),
		OperationID: operation.ID,
		Level:       operation.Level,
		Address:     address,
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
		diffs := prepareBigMapDiffsToEnrich([]bigmapdiff.BigMapDiff{*bmd}, false)
		bm.AddDiffs(diffs...)
		b.temporaryPointers[bmd.Ptr] = bm
		return nil
	}

	return errors.Wrapf(ErrUnknownTemporaryPointer, "%d", ptr)
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

func (b *Babylon) getCopyBigMapDiff(src int64, address string) (bmd []bigmapdiff.BigMapDiff, err error) {
	if src > -1 {
		states, err := b.bigmapdiffs.GetByPtr(address, src)
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
