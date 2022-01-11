package handlers

import (
	"net/http"

	"github.com/baking-bad/bcdhub/internal/models/contract"
	"github.com/baking-bad/bcdhub/internal/models/types"
	"github.com/gin-gonic/gin"
)

// GetContract godoc
// @Summary Get contract info
// @Description Get full contract info
// @Tags contract
// @ID get-contract
// @Param network path string true "Network"
// @Param address path string true "KT address" minlength(36) maxlength(36)
// @Accept  json
// @Produce  json
// @Success 200 {object} Contract
// @Success 204 {object} gin.H
// @Failure 400 {object} Error
// @Failure 404 {object} Error
// @Failure 500 {object} Error
// @Router /v1/contract/{network}/{address} [get]
func (ctx *Context) GetContract(c *gin.Context) {
	var req getContractRequest
	if err := c.BindUri(&req); ctx.handleError(c, err, http.StatusNotFound) {
		return
	}

	contract, err := ctx.Contracts.Get(req.NetworkID(), req.Address)
	if err != nil {
		if ctx.Storage.IsRecordNotFound(err) {
			c.SecureJSON(http.StatusNoContent, gin.H{})
			return
		}
		ctx.handleError(c, err, 0)
		return
	}

	res, err := ctx.contractPostprocessing(contract)
	if ctx.handleError(c, err, 0) {
		return
	}
	c.SecureJSON(http.StatusOK, res)
}

// GetRandomContract godoc
// @Summary Show random contract
// @Description Get random contract with 2 or more operations
// @Tags contract
// @ID get-random-contract
// @Param network query string false "Network"
// @Accept  json
// @Produce  json
// @Success 200 {object} Contract
// @Success 204 {object} gin.H
// @Failure 500 {object} Error
// @Router /v1/pick_random [get]
func (ctx *Context) GetRandomContract(c *gin.Context) {
	var req networkQueryRequest
	if err := c.BindQuery(&req); ctx.handleError(c, err, http.StatusBadRequest) {
		return
	}
	networks := make([]types.Network, 0)

	network := req.NetworkID()
	if network != types.Empty {
		networks = append(networks, network)
	} else {
		for i := range ctx.Config.API.Networks {
			if net := types.NewNetwork(ctx.Config.API.Networks[i]); net != types.Empty {
				networks = append(networks, net)
			}
		}
	}

	contract, err := ctx.Contracts.GetRandom(networks...)
	if err != nil {
		if ctx.Storage.IsRecordNotFound(err) {
			c.SecureJSON(http.StatusNoContent, gin.H{})
			return
		}
		ctx.handleError(c, err, 0)
		return
	}

	res, err := ctx.contractPostprocessing(contract)
	if ctx.handleError(c, err, 0) {
		return
	}
	c.SecureJSON(http.StatusOK, res)
}

func (ctx *Context) contractPostprocessing(contract contract.Contract) (Contract, error) {
	var res Contract
	res.FromModel(contract)

	res.Alias = ctx.CachedAlias(contract.Network, contract.Address)
	res.DelegateAlias = ctx.CachedAlias(contract.Network, contract.Delegate)

	if alias, err := ctx.TZIP.Get(contract.Network, contract.Address); err == nil {
		res.Slug = alias.Slug
	} else if !ctx.Storage.IsRecordNotFound(err) {
		return res, err
	}

	stats, err := ctx.Contracts.Stats(contract)
	if err != nil {
		return res, err
	}
	res.SameCount = stats.SameCount
	res.SimilarCount = stats.SimilarCount

	return res, nil
}
