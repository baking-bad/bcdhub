package handlers

import (
	"fmt"
	"net/http"

	"github.com/baking-bad/bcdhub/internal/bcd"
	"github.com/baking-bad/bcdhub/internal/bcd/ast"
	"github.com/baking-bad/bcdhub/internal/bcd/consts"
	"github.com/baking-bad/bcdhub/internal/bcd/formatter"
	"github.com/baking-bad/bcdhub/internal/config"
	"github.com/baking-bad/bcdhub/internal/models/bigmapaction"
	"github.com/baking-bad/bcdhub/internal/models/bigmapdiff"
	"github.com/baking-bad/bcdhub/internal/models/types"
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
)

// GetBigMap godoc
// @Summary Get big map info by pointer
// @Description Get big map info by pointer
// @Tags bigmap
// @ID get-bigmap
// @Param network path string true "Network"
// @Param ptr path integer true "Big map pointer"
// @Accept  json
// @Produce  json
// @Success 200 {object} GetBigMapResponse
// @Failure 400 {object} Error
// @Failure 500 {object} Error
// @Router /v1/bigmap/{network}/{ptr} [get]
func GetBigMap() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.MustGet("context").(*config.Context)

		var req getBigMapRequest
		if err := c.BindUri(&req); handleError(c, ctx.Storage, err, http.StatusBadRequest) {
			return
		}

		stats, err := ctx.BigMapDiffs.GetStats(req.Ptr)
		if handleError(c, ctx.Storage, err, 0) {
			return
		}

		res := GetBigMapResponse{
			Network:    req.Network,
			Ptr:        req.Ptr,
			Address:    stats.Contract,
			TotalKeys:  uint(stats.Total),
			ActiveKeys: uint(stats.Active),
		}

		if stats.Total == 0 {
			actions, err := ctx.BigMapActions.Get(req.Ptr, 1, 0)
			if handleError(c, ctx.Storage, err, 0) {
				return
			}
			if len(actions) > 0 {
				res.Address = actions[0].Address
			}
		} else {
			destination, err := ctx.Accounts.Get(res.Address)
			if handleError(c, ctx.Storage, err, 0) {
				return
			}
			res.ContractAlias = destination.Alias

			operation, err := ctx.Operations.Last(
				map[string]interface{}{
					"status":         types.OperationStatusApplied,
					"destination_id": destination.ID,
				}, 0)
			if handleError(c, ctx.Storage, err, 0) {
				return
			}
			proto, err := ctx.Cache.ProtocolByID(operation.ProtocolID)
			if handleError(c, ctx.Storage, err, 0) {
				return
			}

			var deffatedStorage []byte
			if proto.SymLink == bcd.SymLinkAlpha {
				deffatedStorage, err = ctx.RPC.GetScriptStorageRaw(c, res.Address, 0)
				if handleError(c, ctx.Storage, err, 0) {
					return
				}
			} else {
				deffatedStorage = operation.DeffatedStorage
			}

			var data ast.UntypedAST
			if err := json.Unmarshal(deffatedStorage, &data); handleError(c, ctx.Storage, err, 0) {
				return
			}

			script, err := ctx.Contracts.ScriptPart(res.Address, proto.SymLink, consts.STORAGE)
			if handleError(c, ctx.Storage, err, 0) {
				return
			}
			storage, err := ast.NewTypedAstFromBytes(script)
			if handleError(c, ctx.Storage, err, 0) {
				return
			}

			if err := storage.Settle(data); handleError(c, ctx.Storage, err, 0) {
				return
			}

			bigMap := storage.FindBigMapByPtr()
			for p, b := range bigMap {
				if p == req.Ptr {
					res.Typedef, _, err = b.Docs(ast.DocsFull)
					if handleError(c, ctx.Storage, err, 0) {
						return
					}
					break
				}
			}
		}

		c.SecureJSON(http.StatusOK, res)
	}
}

// GetBigMapHistory godoc
// @Summary Get big map actions (alloc/copy/remove)
// @Description Get big map actions (alloc/copy/remove)
// @Tags bigmap
// @ID get-bigmap-history
// @Param network path string true "Network"
// @Param ptr path integer true "Big map pointer"
// @Accept  json
// @Produce  json
// @Success 200 {object} BigMapHistoryResponse
// @Success 204 {object} gin.H
// @Failure 400 {object} Error
// @Failure 500 {object} Error
// @Router /v1/bigmap/{network}/{ptr}/history [get]
func GetBigMapHistory() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.MustGet("context").(*config.Context)

		var req getBigMapRequest
		if err := c.BindUri(&req); handleError(c, ctx.Storage, err, http.StatusBadRequest) {
			return
		}

		bm, err := ctx.BigMapActions.Get(req.Ptr, 0, 0)
		if handleError(c, ctx.Storage, err, 0) {
			return
		}
		if bm == nil {
			c.SecureJSON(http.StatusNoContent, gin.H{})
			return
		}

		c.SecureJSON(http.StatusOK, prepareBigMapHistory(bm, req.Ptr))
	}
}

// GetBigMapKeys godoc
// @Summary Get big map keys by pointer
// @Description Get big map keys by pointer
// @Tags bigmap
// @ID get-bigmap-keys
// @Param network path string true "Network"
// @Param ptr path integer true "Big map pointer"
// @Param q query string false "Search string"
// @Param offset query integer false "Offset"
// @Param size query integer false "Requested count" mininum(1) maximum(10)
// @Param max_level query integer false "Max level filter" minimum(0)
// @Param min_level query integer false "Min level filter" minimum(0)
// @Accept json
// @Produce json
// @Success 200 {array} BigMapResponseItem
// @Failure 400 {object} Error
// @Failure 500 {object} Error
// @Router /v1/bigmap/{network}/{ptr}/keys [get]
func GetBigMapKeys() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.MustGet("context").(*config.Context)

		var req getBigMapRequest
		if err := c.BindUri(&req); handleError(c, ctx.Storage, err, http.StatusBadRequest) {
			return
		}
		var pageReq bigMapSearchRequest
		if err := c.BindQuery(&pageReq); handleError(c, ctx.Storage, err, http.StatusBadRequest) {
			return
		}

		var states []bigmapdiff.BigMapState
		keys, err := ctx.BigMapDiffs.Keys(bigmapdiff.GetContext{
			Ptr:      &req.Ptr,
			Size:     pageReq.Size,
			Offset:   pageReq.Offset,
			MaxLevel: pageReq.MaxLevel,
			MinLevel: pageReq.MinLevel,
		})

		if handleError(c, ctx.Storage, err, 0) {
			return
		}
		states = keys

		symLink, err := getCurrentSymLink(ctx.Blocks)
		if handleError(c, ctx.Storage, err, 0) {
			return
		}

		response, err := prepareBigMapKeys(ctx, states, symLink)
		if handleError(c, ctx.Storage, err, 0) {
			return
		}

		c.SecureJSON(http.StatusOK, response)
	}
}

// GetBigMapByKeyHash godoc
// @Summary Get big map diffs by pointer and key hash
// @Description Get big map diffs by pointer and key hash
// @Tags bigmap
// @ID get-bigmap-keyhash
// @Param network path string true "Network"
// @Param ptr path integer true "Big map pointer"
// @Param key_hash path string true "Key hash in big map" minlength(54) maxlength(54)
// @Param offset query integer false "Offset"
// @Param size query integer false "Requested count" mininum(1) maximum(10)
// @Accept json
// @Produce json
// @Success 200 {object} BigMapDiffByKeyResponse
// @Success 204 {object} gin.H
// @Failure 400 {object} Error
// @Failure 500 {object} Error
// @Router /v1/bigmap/{network}/{ptr}/keys/{key_hash} [get]
func GetBigMapByKeyHash() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.MustGet("context").(*config.Context)

		var req getBigMapByKeyHashRequest
		if err := c.BindUri(&req); handleError(c, ctx.Storage, err, http.StatusBadRequest) {
			return
		}

		var pageReq pageableRequest
		if err := c.BindQuery(&pageReq); handleError(c, ctx.Storage, err, http.StatusBadRequest) {
			return
		}

		bm, total, err := ctx.BigMapDiffs.GetByPtrAndKeyHash(req.Ptr, req.KeyHash, pageReq.Size, pageReq.Offset)
		if handleError(c, ctx.Storage, err, 0) {
			return
		}

		if total == 0 {
			c.SecureJSON(http.StatusNoContent, gin.H{})
			return
		}
		symLink, err := getCurrentSymLink(ctx.Blocks)
		if handleError(c, ctx.Storage, err, 0) {
			return
		}

		response, err := prepareBigMapItem(ctx, bm, req.KeyHash, symLink)
		if handleError(c, ctx.Storage, err, 0) {
			return
		}

		response.Total = total
		c.SecureJSON(http.StatusOK, response)
	}
}

// GetCurrentBigMapKeyHash godoc
// @Summary Get current big map value by pointer and key hash
// @Description Get current big map value by pointer and key hash
// @Tags bigmap
// @ID get-bigmap-keyhash-current
// @Param network path string true "Network"
// @Param ptr path integer true "Big map pointer"
// @Param key_hash path string true "Key hash in big map" minlength(54) maxlength(54)
// @Accept json
// @Produce json
// @Success 200 {object} BigMapDiffByKeyResponse
// @Success 404 {object} Error
// @Failure 400 {object} Error
// @Failure 500 {object} Error
// @Router /v1/bigmap/{network}/{ptr}/keys/{key_hash}/current [get]
func GetCurrentBigMapKeyHash() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.MustGet("context").(*config.Context)

		var req getBigMapByKeyHashRequest
		if err := c.BindUri(&req); handleError(c, ctx.Storage, err, http.StatusBadRequest) {
			return
		}

		state, err := ctx.BigMapDiffs.Current(req.KeyHash, req.Ptr)
		if handleError(c, ctx.Storage, err, 0) {
			return
		}

		symLink, err := getCurrentSymLink(ctx.Blocks)
		if handleError(c, ctx.Storage, err, 0) {
			return
		}

		bigMapType, err := getBigMapType(ctx, state.Contract, state.Ptr, symLink)
		if handleError(c, ctx.Storage, err, 0) {
			return
		}

		key, value, keyString, err := prepareItem(state.Key, state.Value, bigMapType)
		if handleError(c, ctx.Storage, err, 0) {
			return
		}

		c.SecureJSON(http.StatusOK, BigMapKeyStateResponse{
			Key:             key,
			KeyString:       keyString,
			Value:           value,
			Removed:         state.Removed,
			KeyHash:         state.KeyHash,
			UpdatesCount:    state.Count,
			LastUpdateLevel: state.LastUpdateLevel,
			LastUpdateTime:  state.LastUpdateTime,
		})
	}
}

// GetBigMapDiffCount godoc
// @Summary Get big map diffs count info by pointer
// @Description Get big map diffs count info by pointer
// @Tags bigmap
// @ID get-bigmapdiff-count
// @Param network path string true "Network"
// @Param ptr path integer true "Big map pointer"
// @Accept  json
// @Produce  json
// @Success 200 {object} CountResponse
// @Failure 400 {object} Error
// @Failure 500 {object} Error
// @Router /v1/bigmap/{network}/{ptr}/count [get]
func GetBigMapDiffCount() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.MustGet("context").(*config.Context)
		var req getBigMapRequest
		if err := c.BindUri(&req); handleError(c, ctx.Storage, err, http.StatusBadRequest) {
			return
		}

		count, err := ctx.BigMapDiffs.Count(req.Ptr)
		if err != nil {
			if ctx.Storage.IsRecordNotFound(err) {
				c.SecureJSON(http.StatusOK, CountResponse{})
				return
			}
			handleError(c, ctx.Storage, err, 0)
			return
		}
		c.SecureJSON(http.StatusOK, CountResponse{count})
	}
}

func prepareBigMapKeys(ctx *config.Context, data []bigmapdiff.BigMapState, symLink string) ([]BigMapResponseItem, error) {
	if len(data) == 0 {
		return []BigMapResponseItem{}, nil
	}

	bigMapType, err := getBigMapType(ctx, data[0].Contract, data[0].Ptr, symLink)
	if err != nil {
		return nil, err
	}

	res := make([]BigMapResponseItem, len(data))
	for i := range data {
		key, value, keyString, err := prepareItem(data[i].Key, data[i].Value, bigMapType)
		if err != nil {
			return nil, err
		}

		res[i] = BigMapResponseItem{
			Item: BigMapItem{
				Key:       key,
				KeyHash:   data[i].KeyHash,
				KeyString: keyString,
				Level:     data[i].LastUpdateLevel,
				Value:     value,
				Timestamp: data[i].LastUpdateTime,
			},
			Count: data[i].Count,
		}
	}
	return res, nil
}

func prepareBigMapItem(ctx *config.Context, data []bigmapdiff.BigMapDiff, keyHash, symLink string) (res BigMapDiffByKeyResponse, err error) {
	if len(data) == 0 {
		return
	}

	bigMapType, err := getBigMapType(ctx, data[0].Contract, data[0].Ptr, symLink)
	if err != nil {
		return
	}

	var key, value interface{}
	values := make([]BigMapDiffItem, len(data))
	for i := range data {
		key, value, _, err = prepareItem(data[i].Key, data[i].Value, bigMapType)
		if err != nil {
			return
		}

		values[i] = BigMapDiffItem{
			Level:     data[i].Level,
			Value:     value,
			Timestamp: data[i].Timestamp,
		}

	}
	res.Values = values
	res.KeyHash = keyHash
	res.Key = key
	return
}

func getBigMapType(ctx *config.Context, contract string, ptr int64, symLink string) (*ast.BigMap, error) {
	storageType, err := getStorageType(ctx.Contracts, contract, symLink)
	if err != nil {
		return nil, err
	}

	actions, err := ctx.BigMapActions.Get(ptr, 2, 0)
	if err != nil {
		return nil, err
	}

	var deffatedStorage []byte
	if len(actions) != 0 {
		operationID := actions[0].OperationID
		if actions[0].Action == types.BigMapActionRemove {
			if len(actions) >= 2 {
				operationID = actions[1].OperationID
			}
		}
		operation, err := ctx.Operations.GetByID(operationID)
		if err != nil {
			return nil, err
		}
		deffatedStorage = operation.DeffatedStorage
	} else {
		contract, err := ctx.Accounts.Get(contract)
		if err != nil {
			return nil, err
		}

		operation, err := ctx.Operations.Last(map[string]interface{}{
			"destination_id": contract.ID,
			"status":         types.OperationStatusApplied,
		}, 0)
		if err != nil {
			return nil, err
		}
		deffatedStorage = operation.DeffatedStorage
	}

	if err := storageType.SettleFromBytes(deffatedStorage); err != nil {
		return nil, err
	}
	bigMaps := storageType.FindBigMapByPtr()
	bigMapType, ok := bigMaps[ptr]
	if !ok {
		return nil, errors.Errorf("can't find ptr in storage: %d", ptr)
	}
	return bigMapType, nil
}

func prepareItem(itemKey, itemValue types.Bytes, bigMapType *ast.BigMap) (key, value *ast.MiguelNode, keyString string, err error) {
	if itemKey != nil {
		keyType := ast.Copy(bigMapType.KeyType)
		key, err = createMiguelForType(keyType, itemKey)
		if err != nil {
			return nil, nil, "", err
		}

		if key.Value != nil {
			switch t := key.Value.(type) {
			case string:
				keyString = t
			case int64:
				keyString = fmt.Sprintf("%d", t)
			default:
				keyString = fmt.Sprintf("%v", t)
			}
		} else {
			keyString, err = formatter.MichelineToMichelsonInline(string(itemKey))
			if err != nil {
				return nil, nil, "", err
			}
		}
	}

	if itemValue != nil {
		valueType := ast.Copy(bigMapType.ValueType)
		valueMiguel, err := createMiguelForType(valueType, itemValue)
		if err != nil {
			return nil, nil, "", err
		}
		value = valueMiguel
	}

	return
}

func createMiguelForType(typ ast.Node, raw []byte) (*ast.MiguelNode, error) {
	var data ast.UntypedAST
	if err := json.Unmarshal(raw, &data); err != nil {
		return nil, err
	}
	if err := typ.ParseValue(data[0]); err != nil {
		return nil, err
	}
	return typ.ToMiguel()
}

func prepareBigMapHistory(arr []bigmapaction.BigMapAction, ptr int64) BigMapHistoryResponse {
	if len(arr) == 0 {
		return BigMapHistoryResponse{}
	}
	response := BigMapHistoryResponse{
		Address: arr[0].Address,
		Ptr:     ptr,
		Items:   make([]BigMapHistoryItem, len(arr)),
	}

	for i := range arr {
		response.Items[i] = BigMapHistoryItem{
			Action:    arr[i].Action.String(),
			Timestamp: arr[i].Timestamp,
		}
		if arr[i].DestinationPtr != nil && *arr[i].DestinationPtr != ptr {
			response.Items[i].DestinationPtr = arr[i].DestinationPtr
		} else if arr[i].SourcePtr != nil && *arr[i].SourcePtr != ptr {
			response.Items[i].SourcePtr = arr[i].SourcePtr
		}
	}

	return response
}
