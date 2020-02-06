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

	res := map[string]interface{}{}
	v, err := ctx.ES.FindSameContracts(req.Address, h, 5)
	if err != nil {
		_ = c.AbortWithError(http.StatusBadRequest, err)
		return
	}
	res["same"] = v

	s, err := ctx.ES.FindSimilarContracts(h, 5)
	if err != nil {
		_ = c.AbortWithError(http.StatusBadRequest, err)
		return
	}
	res["similar"] = s
	c.JSON(http.StatusOK, res)
}
