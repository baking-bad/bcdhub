package handlers

import (
	"net/http"

	"github.com/aopoltorzhicky/bcdhub/internal/elastic"
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
	similar, err := ctx.getSimilarDiffs(s, req.Address, req.Network)
	if err != nil {
		_ = c.AbortWithError(http.StatusBadRequest, err)
		return
	}
	res["similar"] = similar
	c.JSON(http.StatusOK, res)
}

func (ctx *Context) getSimilarDiffs(similar []elastic.SimilarContract, address, network string) ([]elastic.SimilarContract, error) {
	for i := 0; i < len(similar)-1; i++ {
		src := &similar[i]
		d, err := ctx.getDiff(address, network, src.Address, src.Network, 0, 0)
		if err != nil {
			return nil, err
		}
		src.Added = d.Added
		src.Removed = d.Removed
	}
	return similar, nil
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
