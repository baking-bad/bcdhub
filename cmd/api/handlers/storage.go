package handlers

import (
	"net/http"

	"github.com/baking-bad/bcdhub/internal/bcd/consts"
	"github.com/baking-bad/bcdhub/internal/bcd/formatter"
	"github.com/baking-bad/bcdhub/internal/contractparser/docstring"
	"github.com/baking-bad/bcdhub/internal/contractparser/meta"
	"github.com/baking-bad/bcdhub/internal/contractparser/newmiguel"
	"github.com/baking-bad/bcdhub/internal/jsonschema"
	"github.com/baking-bad/bcdhub/internal/models/operation"
	"github.com/gin-gonic/gin"
	"github.com/tidwall/gjson"
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
// @Success 200 {object} newmiguel.Node
// @Failure 400 {object} Error
// @Failure 500 {object} Error
// @Router /v1/contract/{network}/{address}/storage [get]
func (ctx *Context) GetContractStorage(c *gin.Context) {
	var req getContractRequest
	if err := c.BindUri(&req); ctx.handleError(c, err, http.StatusBadRequest) {
		return
	}

	var sReq storageRequest
	if err := c.BindQuery(&sReq); ctx.handleError(c, err, http.StatusBadRequest) {
		return
	}
	rpc, err := ctx.GetRPC(req.Network)
	if ctx.handleError(c, err, http.StatusBadRequest) {
		return
	}
	if sReq.Level == 0 {
		block, err := ctx.Blocks.Last(req.Network)
		if ctx.handleError(c, err, 0) {
			return
		}
		sReq.Level = int(block.Level)
	}

	deffatedStorage, err := rpc.GetScriptStorageJSON(req.Address, int64(sReq.Level))
	if ctx.handleError(c, err, 0) {
		return
	}
	header, err := rpc.GetHeader(int64(sReq.Level))
	if ctx.handleError(c, err, 0) {
		return
	}

	metadata, err := meta.GetSchema(ctx.Schema, req.Address, consts.STORAGE, header.Protocol)
	if ctx.handleError(c, err, 0) {
		return
	}
	resp, err := newmiguel.MichelineToMiguel(deffatedStorage, metadata)
	if ctx.handleError(c, err, 0) {
		return
	}

	c.JSON(http.StatusOK, resp)
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
// @Failure 500 {object} Error
// @Router /v1/contract/{network}/{address}/storage/raw [get]
func (ctx *Context) GetContractStorageRaw(c *gin.Context) {
	var req getContractRequest
	if err := c.BindUri(&req); ctx.handleError(c, err, http.StatusBadRequest) {
		return
	}
	var sReq storageRequest
	if err := c.BindQuery(&sReq); ctx.handleError(c, err, http.StatusBadRequest) {
		return
	}
	filters := map[string]interface{}{
		"destination": req.Address,
		"network":     req.Network,
	}
	if sReq.Level > 0 {
		filters["level"] = sReq.Level
	}

	ops, err := ctx.Operations.Get(filters, 1, true)
	if ctx.handleError(c, err, 0) {
		return
	}
	if len(ops) == 0 {
		c.JSON(http.StatusNoContent, "")
		return
	}

	s := gjson.Parse(ops[0].DeffatedStorage)
	resp, err := formatter.MichelineToMichelson(s, false, formatter.DefLineSize)
	if ctx.handleError(c, err, 0) {
		return
	}

	c.JSON(http.StatusOK, resp)
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
// @Failure 500 {object} Error
// @Router /v1/contract/{network}/{address}/storage/rich [get]
func (ctx *Context) GetContractStorageRich(c *gin.Context) {
	var req getContractRequest
	if err := c.BindUri(&req); ctx.handleError(c, err, http.StatusBadRequest) {
		return
	}
	var sReq storageRequest
	if err := c.BindQuery(&sReq); ctx.handleError(c, err, http.StatusBadRequest) {
		return
	}
	filters := map[string]interface{}{
		"destination": req.Address,
		"network":     req.Network,
	}
	if sReq.Level > 0 {
		filters["level"] = sReq.Level
	}

	ops, err := ctx.Operations.Get(filters, 2, true)
	if ctx.handleError(c, err, 0) {
		return
	}
	if len(ops) == 0 {
		c.JSON(http.StatusNoContent, gin.H{})
		return
	}

	prev := operation.Operation{}
	if len(ops) > 1 {
		prev = ops[1]
	}

	bmd, err := ctx.BigMapDiffs.GetForAddress(req.Address)
	if ctx.handleError(c, err, 0) {
		return
	}

	resp, err := enrichStorage(ops[0].DeffatedStorage, prev.DeffatedStorage, bmd, ops[0].Protocol, true, false)
	if ctx.handleError(c, err, 0) {
		return
	}

	c.JSON(http.StatusOK, resp.Value())
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
// @Failure 500 {object} Error
// @Router /v1/contract/{network}/{address}/storage/schema [get]
func (ctx *Context) GetContractStorageSchema(c *gin.Context) {
	var req getContractRequest
	if err := c.BindUri(&req); ctx.handleError(c, err, http.StatusBadRequest) {
		return
	}
	var ssReq storageSchemaRequest
	if err := c.BindQuery(&ssReq); ctx.handleError(c, err, http.StatusBadRequest) {
		return
	}

	metadata, err := ctx.getStorageMetadata(req.Address, req.Network)
	if ctx.handleError(c, err, 0) {
		return
	}

	schema := new(EntrypointSchema)

	data, err := docstring.GetStorage(metadata)
	if ctx.handleError(c, err, 0) {
		return
	}
	if len(data) > 0 {
		schema.EntrypointType = data[0]
	}
	schema.Schema, err = jsonschema.Create("0", metadata)
	if ctx.handleError(c, err, 0) {
		return
	}

	if ssReq.FillType == "current" {
		rpc, err := ctx.GetRPC(req.Network)
		if ctx.handleError(c, err, 0) {
			return
		}
		storage, err := rpc.GetScriptStorageJSON(req.Address, 0)
		if ctx.handleError(c, err, 0) {
			return
		}
		schema.DefaultModel = make(jsonschema.DefaultModel)
		if err := schema.DefaultModel.Fill(storage, metadata); ctx.handleError(c, err, 0) {
			return
		}
	}

	c.JSON(http.StatusOK, schema)
}
