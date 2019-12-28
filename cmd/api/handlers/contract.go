package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/aopoltorzhicky/bcdhub/internal/db/contract"
)

type getContractRequest struct {
	ID int64 `uri:"id"`
}

type getContractByNetworkAndAddressRequest struct {
	Network string `uri:"network"`
	Address string `uri:"address"`
}

// GetContract -
func (ctx *Context) GetContract(c *gin.Context) {
	var req getContractRequest
	if err := c.BindUri(&req); err != nil {
		_ = c.AbortWithError(http.StatusBadRequest, err)
		return
	}

	var cntr contract.Contract
	if err := ctx.DB.Where("id = ?", req.ID).First(&cntr).Error; err != nil {
		_ = c.AbortWithError(http.StatusBadRequest, err)
		return
	}

	c.JSON(http.StatusOK, cntr)
}

// GetContractByNetworkAndAddress -
func (ctx *Context) GetContractByNetworkAndAddress(c *gin.Context) {
	var req getContractByNetworkAndAddressRequest
	if err := c.BindUri(&req); err != nil {
		_ = c.AbortWithError(http.StatusBadRequest, err)
		return
	}

	var cntr contract.Contract
	if err := ctx.DB.Where("network = ? AND address = ?", req.Network, req.Address).First(&cntr).Error; err != nil {
		_ = c.AbortWithError(http.StatusBadRequest, err)
		return
	}

	c.JSON(http.StatusOK, cntr)
}
