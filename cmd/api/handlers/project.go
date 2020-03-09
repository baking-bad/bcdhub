package handlers

import (
	"net/http"

	"github.com/aopoltorzhicky/bcdhub/internal/elastic"
	"github.com/aopoltorzhicky/bcdhub/internal/models"
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

	s, err := ctx.ES.GetSimilarContracts(contract)
	if handleError(c, err, 0) {
		return
	}
	similar, err := ctx.getSimilarDiffs(s, contract)
	if handleError(c, err, 0) {
		return
	}
	c.JSON(http.StatusOK, similar)
}

func (ctx *Context) getSimilarDiffs(similar []elastic.SimilarContract, contract models.Contract) ([]elastic.SimilarContract, error) {
	if similar == nil {
		return []elastic.SimilarContract{}, nil
	}
	for i := 0; i < len(similar); i++ {
		src := &similar[i]
		d, err := ctx.getDiff(contract.Address, contract.Network, src.Address, src.Network, 0, 0)
		if err != nil {
			return nil, err
		}
		src.Added = d.Added
		src.Removed = d.Removed
		if contract.MedianConsumedGas != 0 {
			src.ConsumedGasDiff = float64((src.MedianConsumedGas - contract.MedianConsumedGas)) / float64(contract.MedianConsumedGas)
		}
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
