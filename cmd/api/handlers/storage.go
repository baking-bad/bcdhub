package handlers

import (
	"net/http"

	"github.com/baking-bad/bcdhub/internal/bcd"
	"github.com/baking-bad/bcdhub/internal/bcd/ast"
	"github.com/baking-bad/bcdhub/internal/bcd/formatter"
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
func (ctx *Context) GetContractStorage(c *gin.Context) {
	var req getContractRequest
	if err := c.BindUri(&req); ctx.handleError(c, err, http.StatusNotFound) {
		return
	}

	var sReq storageRequest
	if err := c.BindQuery(&sReq); ctx.handleError(c, err, http.StatusBadRequest) {
		return
	}

	network := req.NetworkID()

	var header block.Block
	var err error
	if sReq.Level == 0 {
		header, err = ctx.Blocks.Last(network)
		if ctx.handleError(c, err, 0) {
			return
		}
	} else {
		header, err = ctx.Blocks.Get(network, int64(sReq.Level))
		if ctx.handleError(c, err, 0) {
			return
		}
	}

	deffatedStorage, err := ctx.getDeffattedStorage(req.NetworkID(), req.Address, int64(sReq.Level))
	if ctx.handleError(c, err, 0) {
		return
	}
	storageType, err := ctx.getStorageType(network, req.Address, header.Protocol.SymLink)
	if ctx.handleError(c, err, 0) {
		return
	}

	var data ast.UntypedAST
	if err := json.Unmarshal(deffatedStorage, &data); ctx.handleError(c, err, 0) {
		return
	}
	if err := storageType.Settle(data); ctx.handleError(c, err, 0) {
		return
	}

	resp, err := storageType.ToMiguel()
	if ctx.handleError(c, err, 0) {
		return
	}

	c.SecureJSON(http.StatusOK, resp)
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
func (ctx *Context) GetContractStorageRaw(c *gin.Context) {
	var req getContractRequest
	if err := c.BindUri(&req); ctx.handleError(c, err, http.StatusNotFound) {
		return
	}
	var sReq storageRequest
	if err := c.BindQuery(&sReq); ctx.handleError(c, err, http.StatusBadRequest) {
		return
	}
	storage, err := ctx.getDeffattedStorage(req.NetworkID(), req.Address, int64(sReq.Level))
	if ctx.handleError(c, err, 0) {
		return
	}
	resp, err := formatter.MichelineStringToMichelson(string(storage), false, formatter.DefLineSize)
	if ctx.handleError(c, err, 0) {
		return
	}

	c.SecureJSON(http.StatusOK, resp)
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
func (ctx *Context) GetContractStorageRich(c *gin.Context) {
	var req getContractRequest
	if err := c.BindUri(&req); ctx.handleError(c, err, http.StatusNotFound) {
		return
	}
	var sReq storageRequest
	if err := c.BindQuery(&sReq); ctx.handleError(c, err, http.StatusBadRequest) {
		return
	}
	storage, err := ctx.getDeffattedStorage(req.NetworkID(), req.Address, int64(sReq.Level))
	if ctx.handleError(c, err, 0) {
		return
	}
	var symLink string
	if sReq.Level == 0 {
		symLink = bcd.GetCurrentSymLink()
	} else {
		block, err := ctx.Blocks.Get(req.NetworkID(), int64(sReq.Level))
		if ctx.handleError(c, err, 0) {
			return
		}
		symLink = block.Protocol.SymLink
	}

	storageType, err := ctx.getStorageType(req.NetworkID(), req.Address, symLink)
	if ctx.handleError(c, err, 0) {
		return
	}

	states, err := ctx.BigMapDiffs.GetForAddress(req.NetworkID(), req.Address)
	if ctx.handleError(c, err, 0) {
		return
	}

	bmd := make([]bigmapdiff.BigMapDiff, 0, len(states))
	for i := range states {
		bmd = append(bmd, states[i].ToDiff())
	}

	if err := prepareStorage(storageType, storage, bmd); ctx.handleError(c, err, 0) {
		return
	}

	response, err := storageType.Nodes[0].ToBaseNode(false)
	if ctx.handleError(c, err, 0) {
		return
	}

	c.SecureJSON(http.StatusOK, response)
}

// GetContractStorageSchema godoc
// @Summary Get contract storage schema
// @Description Get contract storage schema
// @Tags contract
// @ID get-contract-storage-schema
// @Param network path string true "Network"
// @Param address path string true "KT address" minlength(36) maxlength(36)
// @Param fill_type query string false "Fill storage type" Enums(empty, current)
// @Accept json
// @Produce json
// @Success 200 {object} EntrypointSchema
// @Failure 400 {object} Error
// @Failure 404 {object} Error
// @Failure 500 {object} Error
// @Router /v1/contract/{network}/{address}/storage/schema [get]
func (ctx *Context) GetContractStorageSchema(c *gin.Context) {
	var req getContractRequest
	if err := c.BindUri(&req); ctx.handleError(c, err, http.StatusNotFound) {
		return
	}
	var ssReq storageSchemaRequest
	if err := c.BindQuery(&ssReq); ctx.handleError(c, err, http.StatusBadRequest) {
		return
	}

	storageType, err := ctx.getStorageType(req.NetworkID(), req.Address, bcd.SymLinkBabylon)
	if ctx.handleError(c, err, 0) {
		return
	}

	schema := new(EntrypointSchema)

	data, err := storageType.GetEntrypointsDocs()
	if ctx.handleError(c, err, 0) {
		return
	}
	if len(data) > 0 {
		schema.EntrypointType = data[0]
	}
	schema.Schema, err = storageType.ToJSONSchema()
	if ctx.handleError(c, err, 0) {
		return
	}

	if ssReq.FillType == "current" {
		storage, err := ctx.getDeffattedStorage(req.NetworkID(), req.Address, 0)
		if ctx.handleError(c, err, 0) {
			return
		}

		if err := storageType.SettleFromBytes(storage); ctx.handleError(c, err, 0) {
			return
		}

		schema.DefaultModel = make(ast.JSONModel)
		storageType.GetJSONModel(schema.DefaultModel)
	}

	c.SecureJSON(http.StatusOK, schema)
}

func (ctx *Context) getDeffattedStorage(network types.Network, address string, level int64) ([]byte, error) {
	destination, err := ctx.Accounts.Get(network, address)
	if err != nil {
		return nil, err
	}

	filters := map[string]interface{}{
		"operation.network": network,
		"destination_id":    destination.ID,
		"status":            types.OperationStatusApplied,
	}
	if level > 0 {
		filters["level"] = level
	}
	operation, err := ctx.Operations.Last(filters, 0)
	if err != nil && !ctx.Storage.IsRecordNotFound(err) {
		return nil, err
	}
	if len(operation.DeffatedStorage) == 0 || ctx.Storage.IsRecordNotFound(err) {
		rpc, err := ctx.GetRPC(network)
		if err != nil {
			return nil, err
		}
		return rpc.GetScriptStorageRaw(address, level)
	}
	return operation.DeffatedStorage, nil

}
