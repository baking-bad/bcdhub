package handlers

import (
	"net/http"

	"github.com/baking-bad/bcdhub/internal/contractparser/consts"
	"github.com/baking-bad/bcdhub/internal/contractparser/docstring"
	"github.com/baking-bad/bcdhub/internal/contractparser/formatter"
	"github.com/baking-bad/bcdhub/internal/contractparser/meta"
	"github.com/baking-bad/bcdhub/internal/contractparser/newmiguel"
	"github.com/baking-bad/bcdhub/internal/elastic"
	"github.com/baking-bad/bcdhub/internal/jsonschema"
	"github.com/baking-bad/bcdhub/internal/models"
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
// @Success 204 {object} gin.H
// @Failure 400 {object} Error
// @Failure 500 {object} Error
// @Router /contract/{network}/{address}/storage [get]
func (ctx *Context) GetContractStorage(c *gin.Context) {
	var req getContractRequest
	if err := c.BindUri(&req); handleError(c, err, http.StatusBadRequest) {
		return
	}

	var sReq storageRequest
	if err := c.BindQuery(&sReq); handleError(c, err, http.StatusBadRequest) {
		return
	}

	var protocol string
	var deffatedStorage gjson.Result
	filters := map[string]interface{}{
		"destination": req.Address,
		"network":     req.Network,
		"status":      "applied",
	}
	if sReq.Level > 0 {
		filters["level"] = sReq.Level
	}
	ops, err := ctx.ES.GetOperations(filters, 1, true)
	if err != nil {
		if !elastic.IsRecordNotFound(err) && handleError(c, err, 0) {
			return
		}
		rpc, err := ctx.GetRPC(req.Network)
		if handleError(c, err, http.StatusBadRequest) {
			return
		}

		deffatedStorage, err = rpc.GetScriptStorageJSON(req.Address, int64(sReq.Level))
		if handleError(c, err, 0) {
			return
		}
		header, err := rpc.GetHeader(int64(sReq.Level))
		if handleError(c, err, 0) {
			return
		}
		protocol = header.Protocol
	} else if len(ops) > 0 {
		protocol = ops[0].Protocol
		deffatedStorage = gjson.Parse(ops[0].DeffatedStorage)
	} else {
		c.JSON(http.StatusNoContent, gin.H{})
		return
	}

	metadata, err := meta.GetMetadata(ctx.ES, req.Address, consts.STORAGE, protocol)
	if handleError(c, err, 0) {
		return
	}
	resp, err := newmiguel.MichelineToMiguel(deffatedStorage, metadata)
	if handleError(c, err, 0) {
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
// @Router /contract/{network}/{address}/storage/raw [get]
func (ctx *Context) GetContractStorageRaw(c *gin.Context) {
	var req getContractRequest
	if err := c.BindUri(&req); handleError(c, err, http.StatusBadRequest) {
		return
	}
	var sReq storageRequest
	if err := c.BindQuery(&sReq); handleError(c, err, http.StatusBadRequest) {
		return
	}
	filters := map[string]interface{}{
		"destination": req.Address,
		"network":     req.Network,
	}
	if sReq.Level > 0 {
		filters["level"] = sReq.Level
	}

	ops, err := ctx.ES.GetOperations(filters, 1, true)
	if handleError(c, err, 0) {
		return
	}
	if len(ops) == 0 {
		c.JSON(http.StatusNoContent, "")
		return
	}

	s := gjson.Parse(ops[0].DeffatedStorage)
	resp, err := formatter.MichelineToMichelson(s, false, formatter.DefLineSize)
	if handleError(c, err, 0) {
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
// @Router /contract/{network}/{address}/storage/rich [get]
func (ctx *Context) GetContractStorageRich(c *gin.Context) {
	var req getContractRequest
	if err := c.BindUri(&req); handleError(c, err, http.StatusBadRequest) {
		return
	}
	var sReq storageRequest
	if err := c.BindQuery(&sReq); handleError(c, err, http.StatusBadRequest) {
		return
	}
	filters := map[string]interface{}{
		"destination": req.Address,
		"network":     req.Network,
	}
	if sReq.Level > 0 {
		filters["level"] = sReq.Level
	}

	ops, err := ctx.ES.GetOperations(filters, 2, true)
	if handleError(c, err, 0) {
		return
	}
	if len(ops) == 0 {
		c.JSON(http.StatusNoContent, gin.H{})
		return
	}

	prev := models.Operation{}
	if len(ops) > 1 {
		prev = ops[1]
	}

	bmd, err := ctx.ES.GetBigMapDiffsForAddress(req.Address)
	if handleError(c, err, 0) {
		return
	}

	resp, err := enrichStorage(ops[0].DeffatedStorage, prev.DeffatedStorage, bmd, ops[0].Protocol, true, false)
	if handleError(c, err, 0) {
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
// @Router /contract/{network}/{address}/storage/schema [get]
func (ctx *Context) GetContractStorageSchema(c *gin.Context) {
	var req getContractRequest
	if err := c.BindUri(&req); handleError(c, err, http.StatusBadRequest) {
		return
	}
	var ssReq storageSchemaRequest
	if err := c.BindQuery(&ssReq); handleError(c, err, http.StatusBadRequest) {
		return
	}

	metadata, err := getStorageMetadata(ctx.ES, req.Address, req.Network)
	if handleError(c, err, 0) {
		return
	}

	schema := new(EntrypointSchema)

	data, err := docstring.GetStorage(metadata)
	if handleError(c, err, 0) {
		return
	}
	if len(data) > 0 {
		schema.EntrypointType = data[0]
	}
	schema.Schema, err = jsonschema.Create("0", metadata)
	if handleError(c, err, 0) {
		return
	}

	if ssReq.FillType == "current" {
		rpc, err := ctx.GetRPC(req.Network)
		if handleError(c, err, 0) {
			return
		}
		storage, err := rpc.GetScriptStorageJSON(req.Address, 0)
		if handleError(c, err, 0) {
			return
		}
		schema.DefaultModel = make(jsonschema.DefaultModel)
		if err := schema.DefaultModel.Fill(storage, metadata); handleError(c, err, 0) {
			return
		}
	}

	c.JSON(http.StatusOK, schema)
}
