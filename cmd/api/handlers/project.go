package handlers

import (
	"net/http"

	"github.com/aopoltorzhicky/bcdhub/internal/elastic"
	"github.com/aopoltorzhicky/bcdhub/internal/models"
	"github.com/gin-gonic/gin"
)

type pageableRequest struct {
	Offset int64 `form:"offset"`
}

// GetSameContracts -
func (ctx *Context) GetSameContracts(c *gin.Context) {
	var req getContractRequest
	if err := c.BindUri(&req); err != nil {
		_ = c.AbortWithError(http.StatusBadRequest, err)
		return
	}

	var pageReq pageableRequest
	if err := c.BindQuery(&pageReq); err != nil {
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

	v, err := ctx.ES.GetSameContracts(contract, 0, pageReq.Offset)
	if err != nil {
		_ = c.AbortWithError(http.StatusBadRequest, err)
		return
	}
	c.JSON(http.StatusOK, v)
}

// GetSimilarContracts -
func (ctx *Context) GetSimilarContracts(c *gin.Context) {
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

	s, err := ctx.ES.GetSimilarContracts(contract)
	if err != nil {
		_ = c.AbortWithError(http.StatusBadRequest, err)
		return
	}
	similar, err := ctx.getSimilarDiffs(s, contract)
	if err != nil {
		_ = c.AbortWithError(http.StatusBadRequest, err)
		return
	}
	c.JSON(http.StatusOK, similar)
}

func (ctx *Context) getSimilarDiffs(similar []elastic.SimilarContract, contract models.Contract) ([]elastic.SimilarContract, error) {
	for i := 0; i < len(similar); i++ {
		src := &similar[i]
		d, err := ctx.getDiff(contract.Address, contract.Network, src.Address, src.Network, 0, 0)
		if err != nil {
			return nil, err
		}
		src.Added = d.Added
		src.Removed = d.Removed
		src.ConsumedGasDiff = float64((src.MedianConsumedGas - contract.MedianConsumedGas)) / float64(contract.MedianConsumedGas)
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
