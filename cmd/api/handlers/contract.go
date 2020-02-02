package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type getContractRequest struct {
	Address string `uri:"address"`
	Network string `uri:"network"`
}

// GetContract -
func (ctx *Context) GetContract(c *gin.Context) {
	var req getContractRequest
	if err := c.BindUri(&req); err != nil {
		_ = c.AbortWithError(http.StatusBadRequest, err)
		return
	}

	by := map[string]interface{}{
		"address": req.Address,
		"network": req.Network,
	}
	cntr, err := ctx.ES.GetContract(by)
	if err != nil {
		_ = c.AbortWithError(http.StatusBadRequest, err)
		return
	}

	c.JSON(http.StatusOK, cntr)
}

// GetRandomContract -
func (ctx *Context) GetRandomContract(c *gin.Context) {
	cntr, err := ctx.ES.GetRandomContract()
	if err != nil {
		_ = c.AbortWithError(http.StatusBadRequest, err)
		return
	}

	c.JSON(http.StatusOK, cntr)
}
