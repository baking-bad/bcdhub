package handlers

import (
	"net/http"

	"github.com/baking-bad/bcdhub/internal/contractparser/consts"
	"github.com/baking-bad/bcdhub/internal/contractparser/formatter"
	"github.com/baking-bad/bcdhub/internal/contractparser/meta"
	"github.com/baking-bad/bcdhub/internal/contractparser/newmiguel"
	"github.com/gin-gonic/gin"
	"github.com/tidwall/gjson"
)

// GetContractStorage -
func (ctx *Context) GetContractStorage(c *gin.Context) {
	var req getContractRequest
	if err := c.BindUri(&req); handleError(c, err, http.StatusBadRequest) {
		return
	}

	storage, err := ctx.ES.GetLastStorage(req.Network, req.Address)
	if handleError(c, err, 0) {
		return
	}

	// bmd, err := ctx.ES.GetBigMapDiffsForAddress(req.Address)
	// if handleError(c, err, 0) {
	// 	return
	// }

	protocol := storage.Get("_source.protocol").String()
	s := gjson.Parse(storage.Get("_source.deffated_storage").String())
	// s, err := enrichStorage(storage.Get("_source.deffated_storage").String(), bmd, protocol, true)
	// if handleError(c, err, 0) {
	// 	return
	// }

	metadata, err := meta.GetMetadata(ctx.ES, req.Address, consts.STORAGE, protocol)
	if handleError(c, err, 0) {
		return
	}

	resp, err := newmiguel.MichelineToMiguel(s, metadata)
	if handleError(c, err, 0) {
		return
	}

	c.JSON(http.StatusOK, resp)
}

// GetContractStorageRaw -
func (ctx *Context) GetContractStorageRaw(c *gin.Context) {
	var req getContractRequest
	if err := c.BindUri(&req); handleError(c, err, http.StatusBadRequest) {
		return
	}

	storage, err := ctx.ES.GetLastStorage(req.Network, req.Address)
	if handleError(c, err, 0) {
		return
	}

	s := gjson.Parse(storage.Get("_source.deffated_storage").String())
	resp, err := formatter.MichelineToMichelson(s, false, formatter.DefLineSize)
	if handleError(c, err, 0) {
		return
	}

	c.JSON(http.StatusOK, resp)
}
