package handlers

import (
	"net/http"

	"github.com/baking-bad/bcdhub/internal/bcd/consts"
	"github.com/baking-bad/bcdhub/internal/config"
	"github.com/baking-bad/bcdhub/internal/models/contract"
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

// GetSameContracts godoc
// @Summary Get same contracts
// @Description Get same contracts
// @Tags contract
// @ID get-contract-same
// @Param network path string true "Network"
// @Param address path string true "KT address" minlength(36) maxlength(36)
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

		count, err := ctx.Domains.SameCount(contract)
		if handleError(c, ctx.Storage, err, 0) {
			return
		}

		response := SameContractsResponse{
			Count:     int64(count),
			Contracts: make([]ContractWithStats, 0),
		}

		same, err := ctx.Domains.Same(req.Network, contract, int(page.Size), int(page.Offset))
		if handleError(c, ctx.Storage, err, 0) {
			return
		}

		for i := range same {
			result, err := contractPostprocessing(ctx, same[i].Contract)
			if handleError(c, ctx.Storage, err, 0) {
				return
			}
			result.Network = same[i].Network
			response.Contracts = append(response.Contracts, ContractWithStats{
				Contract:  result,
				SameCount: response.Count,
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
		if res.Alias == "" && contractMetadata.Name != consts.Unknown {
			res.Alias = contractMetadata.Name
		}
	} else if !ctx.Storage.IsRecordNotFound(err) {
		return res, err
	}
	return res, nil
}

func contractWithStatsPostprocessing(ctxs config.Contexts, ctx *config.Context, contractModel contract.Contract) (ContractWithStats, error) {
	c, err := contractPostprocessing(ctx, contractModel)
	if err != nil {
		return ContractWithStats{}, err
	}
	res := ContractWithStats{c, -1}

	stats, err := ctx.Domains.SameCount(contractModel)
	if err != nil {
		return res, err
	}
	res.SameCount += int64(stats)

	for _, cur := range ctxs {
		if cur.Network == ctx.Network {
			continue
		}

		var buf contract.Contract
		switch {
		case contractModel.AlphaID > 0:
			script, err := cur.Scripts.ByHash(contractModel.Alpha.Hash)
			if err != nil {
				if cur.Storage.IsRecordNotFound(err) {
					continue
				}
				return res, err
			}
			buf.AlphaID = script.ID
		case contractModel.BabylonID > 0:
			script, err := cur.Scripts.ByHash(contractModel.Babylon.Hash)
			if err != nil {
				if cur.Storage.IsRecordNotFound(err) {
					continue
				}
				return res, err
			}
			buf.BabylonID = script.ID
		case contractModel.JakartaID > 0:
			script, err := cur.Scripts.ByHash(contractModel.Jakarta.Hash)
			if err != nil {
				if cur.Storage.IsRecordNotFound(err) {
					continue
				}
				return res, err
			}
			buf.JakartaID = script.ID
		}
		stats, err := cur.Domains.SameCount(buf)
		if err != nil {
			return res, err
		}
		res.SameCount += int64(stats)
	}
	return res, nil
}
