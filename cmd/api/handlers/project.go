package handlers

import (
	"net/http"

	"github.com/aopoltorzhicky/bcdhub/internal/elastic"
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
	similar, err := ctx.getSimilarDiffs(s)
	if err != nil {
		_ = c.AbortWithError(http.StatusBadRequest, err)
		return
	}
	res["similar"] = similar
	c.JSON(http.StatusOK, res)
}

func (ctx *Context) getSimilarDiffs(similar []elastic.SimilarContract) ([]elastic.SimilarContract, error) {
	for i := 0; i < len(similar)-1; i++ {
		src := &similar[i]
		dest := similar[i+1]
		d, err := ctx.getDiff(src.Address, src.Network, dest.Address, dest.Network, 0, 0)
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
