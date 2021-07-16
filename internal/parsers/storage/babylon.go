package storage

import (
	stdJSON "encoding/json"

	"github.com/baking-bad/bcdhub/internal/bcd/ast"
	"github.com/baking-bad/bcdhub/internal/models"
	"github.com/baking-bad/bcdhub/internal/models/bigmap"
	"github.com/baking-bad/bcdhub/internal/models/operation"
	"github.com/baking-bad/bcdhub/internal/models/types"
	"github.com/baking-bad/bcdhub/internal/noderpc"
	"github.com/baking-bad/bcdhub/internal/parsers"
	"github.com/pkg/errors"
)

// Babylon -
type Babylon struct {
	bigmapRepo bigmap.Repository
	statesRepo bigmap.StateRepository
	general    models.GeneralRepository
	rpc        noderpc.INode

	ptrMap            map[int64]int64
	temporaryPointers map[int64]*ast.BigMap
}

// NewBabylon -
func NewBabylon(bigmaps bigmap.Repository, statesRepo bigmap.StateRepository, general models.GeneralRepository, rpc noderpc.INode) *Babylon {
	return &Babylon{
		bigmapRepo: bigmaps,
		statesRepo: statesRepo,
		general:    general,
		rpc:        rpc,

		ptrMap:            make(map[int64]int64),
		temporaryPointers: make(map[int64]*ast.BigMap),
	}
}

// ParseTransaction -
func (b *Babylon) ParseTransaction(content noderpc.Operation, operation *operation.Operation) (*parsers.Result, error) {
	result := content.GetResult()
	if result == nil {
		return nil, nil
	}
	operation.DeffatedStorage = result.Storage

	return b.handleBigMapDiff(result, *content.Destination, operation, result.Storage)
}

// ParseOrigination -
func (b *Babylon) ParseOrigination(content noderpc.Operation, operation *operation.Operation) (*parsers.Result, error) {
	result := content.GetResult()
	if result == nil {
		return nil, nil
	}

	var scriptData struct {
		Storage stdJSON.RawMessage `json:"storage"`
	}
	if err := json.Unmarshal(content.Script, &scriptData); err != nil {
		return nil, err
	}

	operation.DeffatedStorage = scriptData.Storage

	return b.handleBigMapDiff(result, result.Originated[0], operation, scriptData.Storage)
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
				b.temporaryPointers[ptr] = typ
				continue
			}
			if ptr < 0 {
				continue
			}

		case bmd.SourceBigMap != nil:
			ptr := *bmd.SourceBigMap
			if typ, ok := bigMaps[ptr]; ok {
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

func (b *Babylon) handleBigMapDiff(result *noderpc.OperationResult, address string, op *operation.Operation, storageData []byte) (*parsers.Result, error) {
	if len(result.BigMapDiffs) == 0 {
		return nil, nil
	}

	if err := b.initPointersTypes(result, op, storageData); err != nil {
		return nil, nil
	}

	res := parsers.NewResult()

	handlers := map[string]func(noderpc.BigMapDiff, string, *operation.Operation, *parsers.Result) error{
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

		if err := handler(result.BigMapDiffs[i], address, op, res); err != nil {
			return nil, err
		}
	}

	for i := range res.BigMaps {
		if bigMap, ok := b.temporaryPointers[res.BigMaps[i].Ptr]; ok {
			res.BigMaps[i].Name = bigMap.TypeName
		}
	}

	return res, nil
}

func (b *Babylon) handleBigMapDiffUpdate(item noderpc.BigMapDiff, address string, operation *operation.Operation, res *parsers.Result) error {
	ptr := *item.BigMap

	bmd := bigmap.Diff{
		BigMap: bigmap.BigMap{
			Ptr:      ptr,
			Contract: address,
			Network:  operation.Network,
		},

		Key:         types.Bytes(item.Key),
		KeyHash:     item.KeyHash,
		OperationID: operation.ID,
		Level:       operation.Level,

		Timestamp:  operation.Timestamp,
		ProtocolID: operation.ProtocolID,
		Value:      types.Bytes(item.Value),
	}

	if err := setBigMapDiffsStrings(&bmd); err != nil {
		return err
	}

	if err := b.addDiff(&bmd, ptr); err != nil {
		return err
	}

	if ptr > -1 {
		operation.BigMapDiffs = append(operation.BigMapDiffs, &bmd)
		res.BigMapState = append(res.BigMapState, bmd.ToState())
	}
	return nil
}

func (b *Babylon) handleBigMapDiffCopy(item noderpc.BigMapDiff, address string, operation *operation.Operation, res *parsers.Result) error {
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

		action, err := b.createBigMapDiffAction("copy", address, &srcPtr, &destinationPtr, res, operation)
		if err != nil {
			return err
		}
		if action != nil {
			operation.BigMapActions = append(operation.BigMapActions, action)
		}
	}

	b.ptrMap[destinationPtr] = sourcePtr

	bmd, err := b.getCopyBigMapDiff(sourcePtr, address, operation.Network)
	if err != nil {
		return err
	}

	b.updateTemporaryPointers(sourcePtr, destinationPtr)

	if len(bmd) > 0 {
		for i := range bmd {
			bmd[i].Level = operation.Level
			bmd[i].Timestamp = operation.Timestamp
			bmd[i].OperationID = operation.ID
			bmd[i].ProtocolID = operation.ProtocolID

			if err := b.addDiff(&bmd[i], destinationPtr); err != nil {
				return err
			}

			if err := setBigMapDiffsStrings(&bmd[i]); err != nil {
				return err
			}

			if destinationPtr > -1 {
				operation.BigMapDiffs = append(operation.BigMapDiffs, &bmd[i])
				res.BigMapState = append(res.BigMapState, bmd[i].ToState())
			}
		}
	}
	return nil
}

func (b *Babylon) handleBigMapDiffRemove(item noderpc.BigMapDiff, address string, operation *operation.Operation, res *parsers.Result) error {
	ptr := *item.BigMap
	if ptr < 0 {
		return nil
	}
	states, err := b.statesRepo.GetByPtr(operation.Network, address, ptr)
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

		if err := setBigMapDiffsStrings(&bmd); err != nil {
			return err
		}

		operation.BigMapDiffs = append(operation.BigMapDiffs, &bmd)
		res.BigMapState = append(res.BigMapState, &states[i])
	}
	action, err := b.createBigMapDiffAction("remove", address, &ptr, nil, res, operation)
	if err != nil {
		return err
	}
	if action != nil {
		operation.BigMapActions = append(operation.BigMapActions, action)
	}
	return nil
}

func (b *Babylon) handleBigMapDiffAlloc(item noderpc.BigMapDiff, address string, operation *operation.Operation, res *parsers.Result) error {
	ptr := *item.BigMap
	if ptr > -1 {
		bm := bigmap.BigMap{
			Network:      operation.Network,
			Ptr:          ptr,
			Contract:     address,
			KeyType:      types.Bytes(item.KeyType),
			ValueType:    types.Bytes(item.ValueType),
			Tags:         types.EmptyTag,
			CreatedLevel: operation.Level,
			CreatedAt:    operation.Timestamp,
		}
		res.BigMaps = append(res.BigMaps, &bm)
		action, err := b.createBigMapDiffAction("alloc", address, &ptr, nil, res, operation)
		if err != nil {
			return err
		}
		if action != nil {
			operation.BigMapActions = append(operation.BigMapActions, action)
		}
	}

	return nil
}

func (b *Babylon) getDiffsFromUpdates(ptr int64) ([]bigmap.Diff, error) {
	bigMap, ok := b.temporaryPointers[ptr]
	if !ok {
		return nil, errors.Wrapf(ErrUnknownTemporaryPointer, "%d", ptr)
	}
	diffs := bigMap.GetDiffs()
	return getBigMapDiffModels(diffs), nil
}

func (b *Babylon) createBigMapDiffAction(action, address string, srcPtr, dstPtr *int64, res *parsers.Result, operation *operation.Operation) (*bigmap.Action, error) {
	bigMapAction := &bigmap.Action{
		Action:      types.NewBigMapAction(action),
		OperationID: operation.ID,
		Level:       operation.Level,
		Timestamp:   operation.Timestamp,
	}

	if srcPtr != nil {
		if *srcPtr > -1 {
			switch bigMapAction.Action {
			case types.BigMapActionAlloc:
				bigMapAction.SourceID = srcPtr
				bigMapAction.Source = bigmap.BigMap{
					Network:      operation.Network,
					Contract:     address,
					Ptr:          *srcPtr,
					CreatedLevel: operation.Level,
					CreatedAt:    operation.Timestamp,
				}
			case types.BigMapActionCopy, types.BigMapActionRemove, types.BigMapActionUpdate:
				srcBigMap, err := b.bigmapRepo.Get(operation.Network, *srcPtr, "")
				if err != nil {
					return nil, err
				}
				bigMapAction.SourceID = &srcBigMap.ID
				bigMapAction.Source = *srcBigMap

				if dstPtr != nil && *dstPtr > -1 {
					destBigMap, err := b.bigmapRepo.Get(operation.Network, *dstPtr, "")
					switch {
					case err == nil:
						bigMapAction.DestinationID = &destBigMap.ID
						bigMapAction.Destination = *destBigMap
					case b.general.IsRecordNotFound(err):
						destBigMap = &bigmap.BigMap{
							Network:      srcBigMap.Network,
							Contract:     address,
							Ptr:          *dstPtr,
							KeyType:      srcBigMap.KeyType,
							ValueType:    srcBigMap.ValueType,
							Tags:         srcBigMap.Tags,
							CreatedLevel: operation.Level,
							CreatedAt:    operation.Timestamp,
						}
						bigMapAction.DestinationID = &destBigMap.ID
						bigMapAction.Destination = *destBigMap

						res.BigMaps = append(res.BigMaps, destBigMap)
					default:
						return nil, err
					}
				}
			default:
				return nil, errors.Errorf("unknown bigmap action: %s", bigMapAction.Action.String())
			}

			return bigMapAction, nil
		}

		if bigMap, ok := b.temporaryPointers[*srcPtr]; ok {
			if bigMapAction.Action == types.BigMapActionCopy {
				bigMapAction.Action = types.BigMapActionAlloc
				keyType, err := json.Marshal(bigMap.KeyType)
				if err != nil {
					return nil, err
				}
				valueType, err := json.Marshal(bigMap.ValueType)
				if err != nil {
					return nil, err
				}
				bm := bigmap.BigMap{
					Network:      operation.Network,
					Contract:     address,
					Ptr:          *dstPtr,
					CreatedLevel: operation.Level,
					CreatedAt:    operation.Timestamp,
					KeyType:      keyType,
					ValueType:    valueType,
				}
				bigMapAction.Source = bm
				bigMapAction.SourceID = &bm.ID
				res.BigMaps = append(res.BigMaps, &bm)
				return bigMapAction, nil
			}
		}
	}

	return nil, nil
}

func (b *Babylon) addDiff(bmd *bigmap.Diff, ptr int64) error {
	if bm, ok := b.temporaryPointers[ptr]; ok {
		return bm.EnrichBigMap(prepareBigMapDiffsToEnrich([]bigmap.Diff{*bmd}, false))
	}

	bigMap, ok := b.temporaryPointers[ptr]
	if !ok {
		return errors.Errorf("Can't find big map pointer: %d", ptr)
	}

	diffs := prepareBigMapDiffsToEnrich([]bigmap.Diff{*bmd}, false)
	bigMap.AddDiffs(diffs...)
	b.temporaryPointers[bmd.BigMap.Ptr] = bigMap
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

func (b *Babylon) getCopyBigMapDiff(src int64, address string, network types.Network) (bmd []bigmap.Diff, err error) {
	if src > -1 {
		states, err := b.statesRepo.GetByPtr(network, address, src)
		if err != nil {
			return nil, err
		}
		bmd = make([]bigmap.Diff, 0, len(states))
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
