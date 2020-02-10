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
	if err := c.BindUri(&req); err != nil {
		_ = c.AbortWithError(http.StatusBadRequest, err)
		return
	}

	storage, err := ctx.ES.GetLastStorage(req.Network, req.Address)
	if err != nil {
		_ = c.AbortWithError(http.StatusBadRequest, err)
		return
	}
	bmd, err := ctx.ES.GetBigMapDiffsForAddress(req.Address)
	if err != nil {
		_ = c.AbortWithError(http.StatusBadRequest, err)
		return
	}

	protocol := storage.Get("_source.protocol").String()
	s, err := enrichStorage(storage.Get("_source.deffated_storage").String(), bmd, protocol)
	if err != nil {
		_ = c.AbortWithError(http.StatusBadRequest, err)
		return
	}

	metadata, err := meta.GetMetadata(ctx.ES, req.Address, req.Network, "storage", protocol)
	if err != nil {
		_ = c.AbortWithError(http.StatusBadRequest, err)
		return
	}

	resp, err := miguel.MichelineToMiguel(s, metadata)
	if err != nil {
		_ = c.AbortWithError(http.StatusBadRequest, err)
		return
	}

	c.JSON(http.StatusOK, resp)
}
