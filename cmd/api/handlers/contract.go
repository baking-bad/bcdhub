package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type getContractRequest struct {
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
