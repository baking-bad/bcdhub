package handlers

import (
	"math/rand"
	"net/http"
	"reflect"

	"github.com/baking-bad/bcdhub/internal/config"
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
// @Success 200 {object} ContractWithStats
// @Success 204 {object} gin.H
// @Failure 400 {object} Error
// @Failure 404 {object} Error
// @Failure 500 {object} Error
// @Router /v1/contract/{network}/{address} [get]
func GetContract() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.MustGet("context").(*config.Context)

		var req getContractRequest
		if err := c.BindUri(&req); handleError(c, ctx.Storage, err, http.StatusNotFound) {
			return
		}

		contract, err := ctx.Contracts.Get(req.Address)
		if err != nil {
			if ctx.Storage.IsRecordNotFound(err) {
				c.SecureJSON(http.StatusNoContent, gin.H{})
				return
			}
			handleError(c, ctx.Storage, err, 0)
			return
		}

		ctxs := c.MustGet("contexts").(config.Contexts)

		res, err := contractWithStatsPostprocessing(ctxs, ctx, contract)
		if handleError(c, ctx.Storage, err, 0) {
			return
		}
		c.SecureJSON(http.StatusOK, res)
	}
}

// GetRandomContract godoc
// @Summary Show random contract
// @Description Get random contract with 2 or more operations
// @Tags contract
// @ID get-random-contract
// @Param network query string false "Network"
// @Accept  json
// @Produce  json
// @Success 200 {object} ContractWithStats
// @Success 204 {object} gin.H
// @Failure 500 {object} Error
// @Router /v1/pick_random [get]
func GetRandomContract() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctxs := c.MustGet("contexts").(config.Contexts)
		anyContext := ctxs.Any()

		var req networkQueryRequest
		if err := c.BindQuery(&req); handleError(c, anyContext.Storage, err, http.StatusBadRequest) {
			return
		}

		network := req.NetworkID()
		if network == types.Empty {
			keys := reflect.ValueOf(ctxs).MapKeys()
			network = keys[rand.Intn(len(keys))].Interface().(types.Network)
		}

		ctx, err := ctxs.Get(network)
		if handleError(c, anyContext.Storage, err, 0) {
			return
		}

		contract, err := ctx.Contracts.GetRandom()
		if err != nil {
			if ctx.Storage.IsRecordNotFound(err) {
				c.SecureJSON(http.StatusNoContent, gin.H{})
				return
			}
			handleError(c, ctx.Storage, err, 0)
			return
		}

		res, err := contractWithStatsPostprocessing(ctxs, ctx, contract)
		if handleError(c, ctx.Storage, err, 0) {
			return
		}
		res.Network = ctx.Network.String()
		c.SecureJSON(http.StatusOK, res)
	}
}

// GetSameContracts godoc
// @Summary Get same contracts
// @Description Get same contracts
// @Tags contract
// @ID get-contract-same
// @Param network path string true "Network"
// @Param address path string true "KT address" minlength(36) maxlength(36)
// @Param manager query string false "Manager"
// @Param offset query integer false "Offset"
// @Param size query integer false "Requested count" mininum(1) maximum(10)
// @Accept json
// @Produce json
// @Success 200 {object} SameContractsResponse
// @Failure 400 {object} Error
// @Failure 404 {object} Error
// @Failure 500 {object} Error
// @Router /v1/contract/{network}/{address}/same [get]
func GetSameContracts() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.MustGet("context").(*config.Context)

		var req getContractRequest
		if err := c.BindUri(&req); handleError(c, ctx.Storage, err, http.StatusNotFound) {
			return
		}

		var page pageableRequest
		if err := c.BindQuery(&page); handleError(c, ctx.Storage, err, http.StatusBadRequest) {
			return
		}

		contract, err := ctx.Contracts.Get(req.Address)
		if handleError(c, ctx.Storage, err, 0) {
			return
		}

		same, err := ctx.Searcher.SameContracts(contract, ctx.Network.String(), ctx.Config.API.Networks, page.Offset, page.Size)
		if handleError(c, ctx.Storage, err, 0) {
			return
		}

		response := SameContractsResponse{
			Count:     same.Count,
			Contracts: make([]ContractWithStats, 0),
		}

		ctxs := c.MustGet("contexts").(config.Contexts)
		for i := range same.Contracts {
			currentContext, ok := ctxs[types.NewNetwork(same.Contracts[i].Network)]
			if !ok {
				continue
			}

			item, err := currentContext.Contracts.Get(same.Contracts[i].Address)
			if handleError(c, ctx.Storage, err, 0) {
				return
			}
			itemContract, err := contractPostprocessing(currentContext, item)
			if handleError(c, ctx.Storage, err, 0) {
				return
			}

			response.Contracts = append(response.Contracts, ContractWithStats{
				Contract:  itemContract,
				SameCount: same.Count,
			})
		}

		c.SecureJSON(http.StatusOK, response)
	}
}

func contractPostprocessing(ctx *config.Context, contract contract.Contract) (Contract, error) {
	var res Contract
	res.FromModel(contract)
	res.Network = ctx.Network.String()

	if contractMetadata, err := ctx.Cache.ContractMetadata(contract.Account.Address); err == nil && contractMetadata != nil {
		res.Slug = contractMetadata.Slug
		if res.Alias == "" {
			res.Alias = contractMetadata.Name
		}
	} else if !ctx.Storage.IsRecordNotFound(err) {
		return res, err
	}
	return res, nil
}

func contractWithStatsPostprocessing(ctxs config.Contexts, ctx *config.Context, contract contract.Contract) (ContractWithStats, error) {
	c, err := contractPostprocessing(ctx, contract)
	if err != nil {
		return ContractWithStats{}, err
	}
	res := ContractWithStats{c, 0}

	for _, cur := range ctxs {
		stats, err := cur.Contracts.Stats(contract)
		if err != nil {
			return res, err
		}
		res.SameCount += int64(stats)
	}
	return res, nil
}
