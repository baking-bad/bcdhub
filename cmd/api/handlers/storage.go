package handlers

import (
	"net/http"
	"time"

	"github.com/baking-bad/bcdhub/internal/contractparser/consts"
	"github.com/baking-bad/bcdhub/internal/contractparser/docstring"
	"github.com/baking-bad/bcdhub/internal/contractparser/formatter"
	"github.com/baking-bad/bcdhub/internal/contractparser/meta"
	"github.com/baking-bad/bcdhub/internal/contractparser/newmiguel"
	"github.com/baking-bad/bcdhub/internal/elastic"
	"github.com/baking-bad/bcdhub/internal/jsonschema"
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
// @Accept json
// @Produce json
// @Success 200 {object} newmiguel.Node
// @Failure 400 {object} Error
// @Failure 500 {object} Error
// @Router /contract/{network}/{address}/storage [get]
func (ctx *Context) GetContractStorage(c *gin.Context) {
	var req getContractRequest
	if err := c.BindUri(&req); handleError(c, err, http.StatusBadRequest) {
		return
	}

	var protocol string
	var deffatedStorage gjson.Result
	last, err := ctx.ES.GetLastOperation(req.Address, req.Network, time.Now().UnixNano()/1000)
	if err != nil {
		if !elastic.IsRecordNotFound(err) && handleError(c, err, 0) {
			return
		}
		rpc, err := ctx.GetRPC(req.Network)
		if handleError(c, err, http.StatusBadRequest) {
			return
		}
		deffatedStorage, err = rpc.GetScriptStorageJSON(req.Address, 0)
		if handleError(c, err, 0) {
			return
		}
		header, err := rpc.GetHeader(0)
		if handleError(c, err, 0) {
			return
		}
		protocol = header.Protocol
	} else {
		protocol = last.Protocol
		deffatedStorage = gjson.Parse(last.DeffatedStorage)
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
// @Accept json
// @Produce json
// @Success 200 {string} string
// @Failure 400 {object} Error
// @Failure 500 {object} Error
// @Router /contract/{network}/{address}/storage/raw [get]
func (ctx *Context) GetContractStorageRaw(c *gin.Context) {
	var req getContractRequest
	if err := c.BindUri(&req); handleError(c, err, http.StatusBadRequest) {
		return
	}

	last, err := ctx.ES.GetLastOperation(req.Address, req.Network, time.Now().UnixNano()/1000)
	if handleError(c, err, 0) {
		return
	}

	s := gjson.Parse(last.DeffatedStorage)
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
// @Accept json
// @Produce json
// @Success 200 {object} gin.H
// @Failure 400 {object} Error
// @Failure 500 {object} Error
// @Router /contract/{network}/{address}/storage/rich [get]
func (ctx *Context) GetContractStorageRich(c *gin.Context) {
	var req getContractRequest
	if err := c.BindUri(&req); handleError(c, err, http.StatusBadRequest) {
		return
	}

	last, err := ctx.ES.GetLastOperation(req.Address, req.Network, time.Now().UnixNano()/1000)
	if handleError(c, err, 0) {
		return
	}

	prev, err := ctx.ES.GetLastOperation(req.Address, req.Network, last.IndexedTime)
	if handleError(c, err, 0) {
		return
	}

	bmd, err := ctx.ES.GetBigMapDiffsForAddress(req.Address)
	if handleError(c, err, 0) {
		return
	}

	resp, err := enrichStorage(last.DeffatedStorage, prev.DeffatedStorage, bmd, last.Protocol, true, false)
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
	schema.Schema, schema.DefaultModel, err = jsonschema.Create("0", metadata)
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
		if err := schema.DefaultModel.Fill(storage, metadata); handleError(c, err, 0) {
			return
		}
	}

	c.JSON(http.StatusOK, schema)
}
