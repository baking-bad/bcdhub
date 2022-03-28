package handlers

import (
	"net/http"
	"strings"

	"github.com/baking-bad/bcdhub/internal/bcd"
	"github.com/baking-bad/bcdhub/internal/config"
	"github.com/baking-bad/bcdhub/internal/models/contract_metadata"
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
)

const (
	maxTokenBalanceBatch = 10
)

// GetInfo godoc
// @Summary Get account info
// @Description Get account info
// @Tags account
// @ID get-account-info
// @Param network path string true "Network"
// @Param address path string true "Address" minlength(36) maxlength(36)
// @Accept  json
// @Produce  json
// @Success 200 {object} AccountInfo
// @Failure 400 {object} Error
// @Failure 404 {object} Error
// @Failure 500 {object} Error
// @Router /v1/account/{network}/{address} [get]
func GetInfo() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.MustGet("context").(*config.Context)

		var req getContractRequest
		if err := c.BindUri(&req); err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, Error{Message: err.Error()})
			return
		}

		acc, err := ctx.Accounts.Get(req.Address)
		if handleError(c, ctx.Storage, err, 0) {
			return
		}
		stats, err := ctx.Statistics.ContractStats(ctx.Network, acc.Address)
		if handleError(c, ctx.Storage, err, 0) {
			return
		}
		block, err := ctx.Blocks.Last()
		if handleError(c, ctx.Storage, err, 0) {
			return
		}
		balance, err := ctx.Cache.TezosBalance(c, acc.Address, block.Level)
		if handleError(c, ctx.Storage, err, 0) {
			return
		}
		c.SecureJSON(http.StatusOK, AccountInfo{
			Address:    acc.Address,
			Alias:      acc.Alias,
			TxCount:    stats.Count,
			Balance:    balance,
			LastAction: stats.LastAction.UTC(),
		})
	}

}

// GetBatchTokenBalances godoc
// @Summary Batch account token balances
// @Description Batch account token balances
// @Tags account
// @ID get-batch-token-balances
// @Param network path string true "Network"
// @Param address query string false "Comma-separated list of addresses (e.g. addr1,addr2,addr3), max 10 addresses"
// @Accept  json
// @Produce  json
// @Success 200 {object} map[string]TokenBalance
// @Failure 400 {object} Error
// @Failure 500 {object} Error
// @Router /v1/account/{network} [get]
func GetBatchTokenBalances() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.MustGet("context").(*config.Context)

		var queryParams batchAddressRequest
		if err := c.BindQuery(&queryParams); handleError(c, ctx.Storage, err, http.StatusBadRequest) {
			return
		}

		accountIDs := make([]int64, 0)
		address := strings.Split(queryParams.Address, ",")
		for i := range address {
			if !bcd.IsAddress(address[i]) {
				handleError(c, ctx.Storage, errors.Errorf("Invalid address: %s", address[i]), http.StatusBadRequest)
				return
			}

			acc, err := ctx.Accounts.Get(address[i])
			if handleError(c, ctx.Storage, err, http.StatusNotFound) {
				return
			}
			accountIDs = append(accountIDs, acc.ID)
		}

		if len(address) > maxTokenBalanceBatch {
			if handleError(c, ctx.Storage, errors.Errorf("Too many addresses: maximum %d allowed", maxTokenBalanceBatch), http.StatusBadRequest) {
				return
			}
		}

		balances, err := ctx.TokenBalances.Batch(accountIDs)
		if handleError(c, ctx.Storage, err, 0) {
			return
		}

		result := make(map[string][]TokenBalance)
		for a, b := range balances {
			result[a] = make([]TokenBalance, len(b))
			for i := range b {
				result[a][i] = TokenBalance{
					Balance: b[i].Balance.String(),
					TokenMetadata: TokenMetadata{
						TokenID:  b[i].TokenID,
						Contract: b[i].Contract,
					},
				}
			}
		}

		c.SecureJSON(http.StatusOK, result)
	}
}

// GetAccountTokenBalances godoc
// @Summary Get account token balances
// @Description Get account token balances
// @Tags account
// @ID get-account-token-balances
// @Param network path string true "Network"
// @Param address path string true "Address" minlength(36) maxlength(36)
// @Param offset query integer false "Offset"
// @Param size query integer false "Requested count" minimum(0) maximum(50)
// @Param contract query string false "Contract address"
// @Param sort_by query string false "Field using for sorting" Enums(token_id, balance)
// @Param hide_empty query string false "Hide zero balances from response" Enums(true, false)
// @Accept  json
// @Produce  json
// @Success 200 {object} TokenBalances
// @Failure 400 {object} Error
// @Failure 404 {object} Error
// @Failure 500 {object} Error
// @Router /v1/account/{network}/{address}/token_balances [get]
func GetAccountTokenBalances() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.MustGet("context").(*config.Context)

		var req getContractRequest
		if err := c.BindUri(&req); handleError(c, ctx.Storage, err, http.StatusNotFound) {
			return
		}
		var queryParams tokenBalanceRequest
		if err := c.BindQuery(&queryParams); handleError(c, ctx.Storage, err, http.StatusBadRequest) {
			return
		}
		balances, err := getAccountBalances(ctx, req.Address, queryParams)
		if handleError(c, ctx.Storage, err, 0) {
			return
		}

		c.SecureJSON(http.StatusOK, balances)
	}
}

func getAccountBalances(ctx *config.Context, address string, req tokenBalanceRequest) (*TokenBalances, error) {
	acc, err := ctx.Accounts.Get(address)
	if err != nil {
		return nil, err
	}

	balances, err := ctx.Domains.TokenBalances(req.Contract, acc.ID, req.Size, req.Offset, req.SortBy, req.HideEmpty)
	if err != nil {
		return nil, err
	}

	response := TokenBalances{
		Balances: make([]TokenBalance, 0),
		Total:    balances.Count,
	}

	for _, token := range balances.Balances {
		tm := TokenMetadataFromElasticModel(token.TokenMetadata, false)
		tb := TokenBalance{
			Balance: token.Balance,
		}
		if !tm.Empty() {
			tb.TokenMetadata = tm
		} else {
			tb.TokenMetadata = TokenMetadata{
				Contract: token.Contract,
				TokenID:  token.TokenID,
			}
		}
		response.Balances = append(response.Balances, tb)
	}

	return &response, nil
}

// GetAccountTokensCountByContract godoc
// @Summary Get account token balances count grouped by contract
// @Description Get account token balances count grouped by contract
// @Tags account
// @ID get-account-token-balances-count
// @Param network path string true "Network"
// @Param address path string true "Address" minlength(36) maxlength(36)
// @Param hide_empty query string false "Hide zero balances from response" Enums(true, false)
// @Accept  json
// @Produce  json
// @Success 200 {object} map[string]int64
// @Failure 400 {object} Error
// @Failure 404 {object} Error
// @Failure 500 {object} Error
// @Router /v1/account/{network}/{address}/count [get]
func GetAccountTokensCountByContract() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.MustGet("context").(*config.Context)

		var req getContractRequest
		if err := c.BindUri(&req); handleError(c, ctx.Storage, err, http.StatusNotFound) {
			return
		}
		var queryParams tokensCountByContractRequest
		if err := c.BindQuery(&queryParams); handleError(c, ctx.Storage, err, http.StatusBadRequest) {
			return
		}
		acc, err := ctx.Accounts.Get(req.Address)
		if handleError(c, ctx.Storage, err, http.StatusNotFound) {
			return
		}
		res, err := ctx.TokenBalances.CountByContract(acc.ID, queryParams.HideEmpty)
		if handleError(c, ctx.Storage, err, 0) {
			return
		}
		c.SecureJSON(http.StatusOK, res)
	}
}

// GetAccountTokensCountByContractWithMetadata godoc
// @Summary Get account token balances count with token metadata grouped by contract
// @Description Get account token balances count with token metadata grouped by contract
// @Tags account
// @ID get-account-token-balances-with-metadata-count
// @Param network path string true "Network"
// @Param address path string true "Address" minlength(36) maxlength(36)
// @Param hide_empty query string false "Hide zero balances from response" Enums(true, false)
// @Accept  json
// @Produce  json
// @Success 200 {object} map[string]TokensCountWithMetadata
// @Failure 400 {object} Error
// @Failure 404 {object} Error
// @Failure 500 {object} Error
// @Router /v1/account/{network}/{address}/count_with_metadata [get]
func GetAccountTokensCountByContractWithMetadata() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.MustGet("context").(*config.Context)

		var req getContractRequest
		if err := c.BindUri(&req); handleError(c, ctx.Storage, err, http.StatusNotFound) {
			return
		}
		var queryParams tokensCountByContractRequest
		if err := c.BindQuery(&queryParams); handleError(c, ctx.Storage, err, http.StatusBadRequest) {
			return
		}

		acc, err := ctx.Accounts.Get(req.Address)
		if handleError(c, ctx.Storage, err, http.StatusNotFound) {
			return
		}

		res, err := ctx.TokenBalances.CountByContract(acc.ID, queryParams.HideEmpty)
		if handleError(c, ctx.Storage, err, 0) {
			return
		}

		response := make(map[string]TokensCountWithMetadata)
		for address, count := range res {
			metadata, err := ctx.Cache.ContractMetadata(address)
			if err != nil {
				if !ctx.Storage.IsRecordNotFound(err) && handleError(c, ctx.Storage, err, 0) {
					return
				} else {
					metadata = &contract_metadata.ContractMetadata{
						Address: address,
					}
				}
			}
			contract, err := ctx.Contracts.Get(metadata.Address)
			if handleError(c, ctx.Storage, err, 0) {
				return
			}
			var t TZIPResponse
			t.FromModel(metadata, false)
			response[address] = TokensCountWithMetadata{
				TZIPResponse: t,
				Count:        count,
				Tags:         contract.Tags.ToArray(),
			}
		}

		c.SecureJSON(http.StatusOK, response)
	}
}

// GetMetadata godoc
// @Summary Get metadata for account
// @Description Returns full metadata for account
// @Tags account
// @ID get-account-tzip
// @Param network path string true "Network"
// @Param address path string true "KT or tz address" minlength(36) maxlength(36)
// @Accept  json
// @Produce  json
// @Success 200 {object} TZIPResponse
// @Success 204 {object} gin.H
// @Failure 400 {object} Error
// @Failure 404 {object} Error
// @Failure 500 {object} Error
// @Router /v1/account/{network}/{address}/metadata [get]
func GetMetadata() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.MustGet("context").(*config.Context)

		var req getContractRequest
		if err := c.BindUri(&req); handleError(c, ctx.Storage, err, http.StatusNotFound) {
			return
		}
		tzip, err := ctx.ContractMetadata.Get(req.Address)
		if err != nil {
			if ctx.Storage.IsRecordNotFound(err) {
				c.SecureJSON(http.StatusNoContent, gin.H{})
			} else {
				handleError(c, ctx.Storage, err, 0)
			}
			return
		}

		var t TZIPResponse
		t.FromModel(tzip, true)
		c.SecureJSON(http.StatusOK, t)
	}
}
