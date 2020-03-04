package handlers

import (
	"net/http"

	"github.com/aopoltorzhicky/bcdhub/internal/contractparser/meta"
	"github.com/aopoltorzhicky/bcdhub/internal/contractparser/miguel"
	"github.com/gin-gonic/gin"
)

// GetContractStorage -
func (ctx *Context) GetContractStorage(c *gin.Context) {
	var req getContractRequest
	if err := c.BindUri(&req); handleError(c, err, http.StatusBadRequest) {
		_ = c.AbortWithError(http.StatusBadRequest, err)
		return
	}

	storage, err := ctx.ES.GetLastStorage(req.Network, req.Address)
	if handleError(c, err, 0) {
		return
	}

	bmd, err := ctx.ES.GetBigMapDiffsForAddress(req.Address)
	if handleError(c, err, 0) {
		return
	}

	protocol := storage.Get("_source.protocol").String()
	s, err := enrichStorage(storage.Get("_source.deffated_storage").String(), bmd, protocol, true)
	if handleError(c, err, 0) {
		return
	}

	metadata, err := meta.GetMetadata(ctx.ES, req.Address, req.Network, "storage", protocol)
	if handleError(c, err, 0) {
		return
	}

	resp, err := miguel.MichelineToMiguel(s, metadata)
	if handleError(c, err, 0) {
		return
	}

	c.JSON(http.StatusOK, resp)
}
