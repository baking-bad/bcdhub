package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type getContractFieldRequest struct {
	getContractRequest
	FieldName string `uri:"field"`
}

// GetContractField -
func (ctx *Context) GetContractField(c *gin.Context) {
	var req getContractFieldRequest
	if err := c.BindUri(&req); err != nil {
		_ = c.AbortWithError(http.StatusBadRequest, err)
		return
	}
	by := map[string]interface{}{
		"address": req.Address,
		"network": req.Network,
	}
	v, err := ctx.ES.GetContractField(by, req.FieldName)
	if err != nil {
		_ = c.AbortWithError(http.StatusBadRequest, err)
		return
	}

	c.JSON(http.StatusOK, v)
}
