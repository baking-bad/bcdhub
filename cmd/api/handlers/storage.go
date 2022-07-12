package handlers

import (
	"net/http"

	"github.com/baking-bad/bcdhub/internal/bcd/ast"
	"github.com/baking-bad/bcdhub/internal/bcd/formatter"
	"github.com/baking-bad/bcdhub/internal/config"
	"github.com/baking-bad/bcdhub/internal/models/bigmapdiff"
	"github.com/baking-bad/bcdhub/internal/models/block"
	"github.com/baking-bad/bcdhub/internal/models/types"
	"github.com/gin-gonic/gin"
)

// GetContractStorage godoc
// @Summary Get contract storage
// @Description Get contract storage
// @Tags contract
// @ID get-contract-storage
// @Param network path string true "Network"
// @Param address path string true "KT address" minlength(36) maxlength(36)
// @Param level query integer false "Level"
// @Accept json
// @Produce json
// @Success 200 {array} ast.MiguelNode
// @Failure 400 {object} Error
// @Failure 404 {object} Error
// @Failure 500 {object} Error
// @Router /v1/contract/{network}/{address}/storage [get]
func GetContractStorage() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.MustGet("context").(*config.Context)
		var req getContractRequest
		if err := c.BindUri(&req); handleError(c, ctx.Storage, err, http.StatusNotFound) {
			return
		}

		var sReq storageRequest
		if err := c.BindQuery(&sReq); handleError(c, ctx.Storage, err, http.StatusBadRequest) {
			return
		}

		var header block.Block
		var err error
		if sReq.Level == 0 {
			header, err = ctx.Blocks.Last()
			if handleError(c, ctx.Storage, err, 0) {
				return
			}
		} else {
			header, err = ctx.Blocks.Get(int64(sReq.Level))
			if handleError(c, ctx.Storage, err, 0) {
				return
			}
		}

		deffatedStorage, err := getDeffattedStorage(c, ctx, req.Address, int64(sReq.Level))
		if handleError(c, ctx.Storage, err, 0) {
			return
		}
		storageType, err := getStorageType(ctx.Contracts, req.Address, header.Protocol.SymLink)
		if handleError(c, ctx.Storage, err, 0) {
			return
		}

		var data ast.UntypedAST
		if err := json.Unmarshal(deffatedStorage, &data); handleError(c, ctx.Storage, err, 0) {
			return
		}
		if err := storageType.Settle(data); handleError(c, ctx.Storage, err, 0) {
			return
		}

		resp, err := storageType.ToMiguel()
		if handleError(c, ctx.Storage, err, 0) {
			return
		}

		c.SecureJSON(http.StatusOK, resp)
	}
}

// GetContractStorageRaw godoc
// @Summary Get contract raw storage
// @Description Get contract raw storage
// @Tags contract
// @ID get-contract-storage-raw
// @Param network path string true "Network"
// @Param address path string true "KT address" minlength(36) maxlength(36)
// @Param level query integer false "Level"
// @Accept json
// @Produce json
// @Success 200 {string} string
// @Success 204 {string} string
// @Failure 400 {object} Error
// @Failure 404 {object} Error
// @Failure 500 {object} Error
// @Router /v1/contract/{network}/{address}/storage/raw [get]
func GetContractStorageRaw() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.MustGet("context").(*config.Context)
		var req getContractRequest
		if err := c.BindUri(&req); handleError(c, ctx.Storage, err, http.StatusNotFound) {
			return
		}
		var sReq storageRequest
		if err := c.BindQuery(&sReq); handleError(c, ctx.Storage, err, http.StatusBadRequest) {
			return
		}
		storage, err := getDeffattedStorage(c, ctx, req.Address, int64(sReq.Level))
		if handleError(c, ctx.Storage, err, 0) {
			return
		}
		resp, err := formatter.MichelineStringToMichelson(string(storage), false, formatter.DefLineSize)
		if handleError(c, ctx.Storage, err, 0) {
			return
		}

		c.SecureJSON(http.StatusOK, resp)
	}
}

// GetContractStorageRich godoc
// @Summary Get contract rich storage
// @Description Get contract rich storage
// @Tags contract
// @ID get-contract-storage-rich
// @Param network path string true "Network"
// @Param address path string true "KT address" minlength(36) maxlength(36)
// @Param level query integer false "Level"
// @Accept json
// @Produce json
// @Success 200 {object} gin.H
// @Success 204 {object} gin.H
// @Failure 400 {object} Error
// @Failure 404 {object} Error
// @Failure 500 {object} Error
// @Router /v1/contract/{network}/{address}/storage/rich [get]
func GetContractStorageRich() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.MustGet("context").(*config.Context)

		var req getContractRequest
		if err := c.BindUri(&req); handleError(c, ctx.Storage, err, http.StatusNotFound) {
			return
		}
		var sReq storageRequest
		if err := c.BindQuery(&sReq); handleError(c, ctx.Storage, err, http.StatusBadRequest) {
			return
		}
		storage, err := getDeffattedStorage(c, ctx, req.Address, int64(sReq.Level))
		if handleError(c, ctx.Storage, err, 0) {
			return
		}

		block, err := ctx.Blocks.Get(int64(sReq.Level))
		if handleError(c, ctx.Storage, err, 0) {
			return
		}

		storageType, err := getStorageType(ctx.Contracts, req.Address, block.Protocol.SymLink)
		if handleError(c, ctx.Storage, err, 0) {
			return
		}

		states, err := ctx.BigMapDiffs.GetForAddress(req.Address)
		if handleError(c, ctx.Storage, err, 0) {
			return
		}

		bmd := make([]bigmapdiff.BigMapDiff, 0, len(states))
		for i := range states {
			bmd = append(bmd, states[i].ToDiff())
		}

		if err := prepareStorage(storageType, storage, bmd); handleError(c, ctx.Storage, err, 0) {
			return
		}

		response, err := storageType.Nodes[0].ToBaseNode(false)
		if handleError(c, ctx.Storage, err, 0) {
			return
		}

		c.SecureJSON(http.StatusOK, response)
	}
}

// GetContractStorageSchema godoc
// @Summary Get contract storage schema
// @Description Get contract storage schema
// @Tags contract
// @ID get-contract-storage-schema
// @Param network path string true "Network"
// @Param address path string true "KT address" minlength(36) maxlength(36)
// @Param fill_type query string false "Fill storage type" Enums(empty, current, initial)
// @Accept json
// @Produce json
// @Success 200 {object} EntrypointSchema
// @Failure 400 {object} Error
// @Failure 404 {object} Error
// @Failure 500 {object} Error
// @Router /v1/contract/{network}/{address}/storage/schema [get]
func GetContractStorageSchema() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.MustGet("context").(*config.Context)
		var req getContractRequest
		if err := c.BindUri(&req); handleError(c, ctx.Storage, err, http.StatusNotFound) {
			return
		}
		var ssReq storageSchemaRequest
		if err := c.BindQuery(&ssReq); handleError(c, ctx.Storage, err, http.StatusBadRequest) {
			return
		}

		symLink, err := getCurrentSymLink(ctx.Blocks)
		if handleError(c, ctx.Storage, err, 0) {
			return
		}

		storageType, err := getStorageType(ctx.Contracts, req.Address, symLink)
		if handleError(c, ctx.Storage, err, 0) {
			return
		}

		schema := new(EntrypointSchema)

		data, err := storageType.GetEntrypointsDocs()
		if handleError(c, ctx.Storage, err, 0) {
			return
		}
		if len(data) > 0 {
			schema.EntrypointType = data[0]
		}
		schema.Schema, err = storageType.ToJSONSchema()
		if handleError(c, ctx.Storage, err, 0) {
			return
		}

		switch ssReq.FillType {
		case "current":
			storage, err := getDeffattedStorage(c, ctx, req.Address, 0)
			if handleError(c, ctx.Storage, err, 0) {
				return
			}

			if err := storageType.SettleFromBytes(storage); handleError(c, ctx.Storage, err, 0) {
				return
			}

			schema.DefaultModel = make(ast.JSONModel)
			storageType.GetJSONModel(schema.DefaultModel)
		case "initial":
			storage, err := getInitialStorage(c, ctx, req.Address)
			if handleError(c, ctx.Storage, err, 0) {
				return
			}

			if err := storageType.SettleFromBytes(storage); handleError(c, ctx.Storage, err, 0) {
				return
			}

			schema.DefaultModel = make(ast.JSONModel)
			storageType.GetJSONModel(schema.DefaultModel)
		default:
		}

		c.SecureJSON(http.StatusOK, schema)
	}
}

func getDeffattedStorage(c *gin.Context, ctx *config.Context, address string, level int64) ([]byte, error) {
	destination, err := ctx.Accounts.Get(address)
	if err != nil {
		return nil, err
	}

	filters := map[string]interface{}{
		"destination_id": destination.ID,
		"status":         types.OperationStatusApplied,
	}
	if level > 0 {
		filters["level"] = level
	}
	operation, err := ctx.Operations.Last(filters, 0)
	if err != nil && !ctx.Storage.IsRecordNotFound(err) {
		return nil, err
	}
	if len(operation.DeffatedStorage) == 0 || ctx.Storage.IsRecordNotFound(err) {
		return ctx.RPC.GetScriptStorageRaw(c, address, level)
	}
	return operation.DeffatedStorage, nil

}

func getInitialStorage(c *gin.Context, ctx *config.Context, address string) ([]byte, error) {
	destination, err := ctx.Accounts.Get(address)
	if err != nil {
		return nil, err
	}

	filters := map[string]interface{}{
		"destination_id": destination.ID,
		"status":         types.OperationStatusApplied,
		"kind":           types.OperationKindOrigination,
	}
	operation, err := ctx.Operations.Last(filters, 0)
	if err != nil && !ctx.Storage.IsRecordNotFound(err) {
		return nil, err
	}
	return operation.DeffatedStorage, nil

}
