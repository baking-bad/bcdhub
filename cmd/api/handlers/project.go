package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type getProjectRequest struct {
	Address string `uri:"address"`
}

// GetProjectContracts -
func (ctx *Context) GetProjectContracts(c *gin.Context) {
	var req getProjectRequest
	if err := c.BindUri(&req); err != nil {
		_ = c.AbortWithError(http.StatusBadRequest, err)
		return
	}
	by := map[string]interface{}{
		"address": req.Address,
	}
	hashCode, err := ctx.ES.GetContractField(by, "hash")
	if err != nil {
		_ = c.AbortWithError(http.StatusBadRequest, err)
		return
	}

	h := make([]string, 0)
	for _, item := range hashCode.([]interface{}) {
		h = append(h, item.(string))
	}

	v, err := ctx.ES.FindProjectContracts(h, 5)
	if err != nil {
		_ = c.AbortWithError(http.StatusBadRequest, err)
		return
	}

	c.JSON(http.StatusOK, v)
}
