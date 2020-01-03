package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// GetProjectContracts -
func (ctx *Context) GetProjectContracts(c *gin.Context) {
	var req getContractRequest
	if err := c.BindUri(&req); err != nil {
		_ = c.AbortWithError(http.StatusBadRequest, err)
		return
	}
	by := map[string]interface{}{
		"address": req.Address,
		"network": req.Network,
	}
	hashCode, err := ctx.ES.GetContractField(by, "hash_code")
	if err != nil {
		_ = c.AbortWithError(http.StatusBadRequest, err)
		return
	}

	v, err := ctx.ES.FindProjectContracts(hashCode.(string), 17.8)
	if err != nil {
		_ = c.AbortWithError(http.StatusBadRequest, err)
		return
	}

	c.JSON(http.StatusOK, v)
}
