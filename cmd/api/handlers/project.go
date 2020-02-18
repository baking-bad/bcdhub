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
	contract, err := ctx.ES.GetContract(by)
	if err != nil {
		_ = c.AbortWithError(http.StatusBadRequest, err)
		return
	}

	res := map[string]interface{}{}
	v, err := ctx.ES.GetSameContracts(contract)
	if err != nil {
		_ = c.AbortWithError(http.StatusBadRequest, err)
		return
	}
	res["same"] = v

	s, err := ctx.ES.GetSimilarContracts(contract)
	if err != nil {
		_ = c.AbortWithError(http.StatusBadRequest, err)
		return
	}
	res["similar"] = s
	c.JSON(http.StatusOK, res)
}

// GetProjects -
func (ctx *Context) GetProjects(c *gin.Context) {
	projects, err := ctx.ES.GetProjectsStats()
	if err != nil {
		_ = c.AbortWithError(http.StatusBadRequest, err)
		return
	}

	c.JSON(http.StatusOK, projects)
}
