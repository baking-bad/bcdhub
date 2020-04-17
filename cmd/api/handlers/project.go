package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// GetSameContracts -
func (ctx *Context) GetSameContracts(c *gin.Context) {
	var req getContractRequest
	if err := c.BindUri(&req); handleError(c, err, http.StatusBadRequest) {
		return
	}

	var pageReq pageableRequest
	if err := c.BindQuery(&pageReq); handleError(c, err, http.StatusBadRequest) {
		return
	}

	by := map[string]interface{}{
		"address": req.Address,
		"network": req.Network,
	}
	contract, err := ctx.ES.GetContract(by)
	if handleError(c, err, 0) {
		return
	}

	v, err := ctx.ES.GetSameContracts(contract, 0, pageReq.Offset)
	if handleError(c, err, 0) {
		return
	}
	c.JSON(http.StatusOK, v)
}

// GetSimilarContracts -
func (ctx *Context) GetSimilarContracts(c *gin.Context) {
	var req getContractRequest
	if err := c.BindUri(&req); handleError(c, err, http.StatusBadRequest) {
		return
	}

	by := map[string]interface{}{
		"address": req.Address,
		"network": req.Network,
	}
	contract, err := ctx.ES.GetContract(by)
	if handleError(c, err, 0) {
		return
	}

	similar, err := ctx.ES.GetSimilarContracts(contract)
	if handleError(c, err, 0) {
		return
	}

	for i := range similar {
		diff, err := ctx.getContractCodeDiff(
			CodeDiffLeg{Address: contract.Address, Network: contract.Network},
			CodeDiffLeg{Address: similar[i].Address, Network: similar[i].Network},
		)
		if handleError(c, err, 0) {
			return
		}
		similar[i].Added = diff.Diff.Added
		similar[i].Removed = diff.Diff.Removed
	}

	c.JSON(http.StatusOK, similar)
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
