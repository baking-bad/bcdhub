package handlers

import (
	"net/http"

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

	level := storage.Get("_source.level").Int()
	s, err := enrichStorage(storage.Get("_source.deffated_storage").String(), bmd, level)
	if err != nil {
		_ = c.AbortWithError(http.StatusBadRequest, err)
		return
	}

	metadata, err := getMetadata(ctx.ES, req.Address, req.Network, "storage", level)
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
