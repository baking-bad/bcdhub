package handlers

import (
	"net/http"
	"strings"

	"github.com/baking-bad/bcdhub/internal/bcd"
	"github.com/baking-bad/bcdhub/internal/models/types"
	"github.com/baking-bad/bcdhub/internal/models/tzip"
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
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
func (ctx *Context) GetInfo(c *gin.Context) {
	var req getContractRequest
	if err := c.BindUri(&req); ctx.handleError(c, err, http.StatusNotFound) {
		return
	}

	stats, err := ctx.Operations.GetStats(req.NetworkID(), req.Address)
	if ctx.handleError(c, err, 0) {
		return
	}
	block, err := ctx.CachedCurrentBlock(req.NetworkID())
	if ctx.handleError(c, err, 0) {
		return
	}

	balance, err := ctx.CachedTezosBalance(req.NetworkID(), req.Address, block.Level)
	if ctx.handleError(c, err, 0) {
		return
	}

	accountInfo := AccountInfo{
		Address:    req.Address,
		Network:    req.Network,
		TxCount:    stats.Count,
		Balance:    balance,
		LastAction: stats.LastAction.UTC(),
	}

	alias, err := ctx.TZIP.Get(req.NetworkID(), req.Address)
	if err != nil {
		if !ctx.Storage.IsRecordNotFound(err) {
			ctx.handleError(c, err, 0)
			return
		}
	} else {
		accountInfo.Alias = alias.Name
	}

	c.SecureJSON(http.StatusOK, accountInfo)
}

// GetBatchTokenBalances godoc
// @Summary Batch account token balances
// @Description Batch account token balances
// @Tags account
// @ID get-batch-token-balances
// @Param network path string true "Network"
// @Param address query string false "Comma-separated list of addresses (e.g. addr1,addr2,addr3)"
// @Accept  json
// @Produce  json
// @Success 200 {object} map[string]TokenBalance
// @Failure 400 {object} Error
// @Failure 500 {object} Error
// @Router /v1/account/{network} [get]
func (ctx *Context) GetBatchTokenBalances(c *gin.Context) {
	var req getByNetwork
	if err := c.BindUri(&req); ctx.handleError(c, err, http.StatusBadRequest) {
		return
	}
	var queryParams batchAddressRequest
	if err := c.BindQuery(&queryParams); ctx.handleError(c, err, http.StatusBadRequest) {
		return
	}
	address := strings.Split(queryParams.Address, ",")
	for i := range address {
		if !bcd.IsAddress(address[i]) {
			ctx.handleError(c, errors.Errorf("Invalid address: %s", address[i]), http.StatusBadRequest)
			return
		}
	}

	balances, err := ctx.TokenBalances.Batch(req.NetworkID(), address)
	if ctx.handleError(c, err, 0) {
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
					Network:  b[i].Network.String(),
				},
			}
		}
	}

	c.SecureJSON(http.StatusOK, result)
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
func (ctx *Context) GetAccountTokenBalances(c *gin.Context) {
	var req getContractRequest
	if err := c.BindUri(&req); ctx.handleError(c, err, http.StatusNotFound) {
		return
	}
	var queryParams tokenBalanceRequest
	if err := c.BindQuery(&queryParams); ctx.handleError(c, err, http.StatusBadRequest) {
		return
	}
	balances, err := ctx.getAccountBalances(req.NetworkID(), req.Address, queryParams)
	if ctx.handleError(c, err, 0) {
		return
	}

	c.SecureJSON(http.StatusOK, balances)
}

func (ctx *Context) getAccountBalances(network types.Network, address string, req tokenBalanceRequest) (*TokenBalances, error) {
	balances, err := ctx.Domains.TokenBalances(network, req.Contract, address, req.Size, req.Offset, req.SortBy, req.HideEmpty)
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
				Network:  token.Network.String(),
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
// @Accept  json
// @Produce  json
// @Success 200 {object} map[string]int64
// @Failure 400 {object} Error
// @Failure 404 {object} Error
// @Failure 500 {object} Error
// @Router /v1/account/{network}/{address}/count [get]
func (ctx *Context) GetAccountTokensCountByContract(c *gin.Context) {
	var req getContractRequest
	if err := c.BindUri(&req); ctx.handleError(c, err, http.StatusNotFound) {
		return
	}
	res, err := ctx.TokenBalances.CountByContract(req.NetworkID(), req.Address)
	if ctx.handleError(c, err, 0) {
		return
	}
	c.SecureJSON(http.StatusOK, res)
}

// GetAccountTokensCountByContractWithMetadata godoc
// @Summary Get account token balances count with token metadata grouped by contract
// @Description Get account token balances count with token metadata grouped by contract
// @Tags account
// @ID get-account-token-balances-with-metadata-count
// @Param network path string true "Network"
// @Param address path string true "Address" minlength(36) maxlength(36)
// @Accept  json
// @Produce  json
// @Success 200 {object} map[string]TokensCountWithMetadata
// @Failure 400 {object} Error
// @Failure 404 {object} Error
// @Failure 500 {object} Error
// @Router /v1/account/{network}/{address}/count_with_metadata [get]
func (ctx *Context) GetAccountTokensCountByContractWithMetadata(c *gin.Context) {
	var req getContractRequest
	if err := c.BindUri(&req); ctx.handleError(c, err, http.StatusNotFound) {
		return
	}
	res, err := ctx.TokenBalances.CountByContract(req.NetworkID(), req.Address)
	if ctx.handleError(c, err, 0) {
		return
	}

	response := make(map[string]TokensCountWithMetadata)
	for address, count := range res {
		metadata, err := ctx.CachedContractMetadata(req.NetworkID(), address)
		if err != nil {
			if !ctx.Storage.IsRecordNotFound(err) && ctx.handleError(c, err, 0) {
				return
			} else {
				metadata = &tzip.TZIP{
					Network: req.NetworkID(),
					Address: address,
				}
			}
		}
		contract, err := ctx.CachedContract(metadata.Network, metadata.Address)
		if ctx.handleError(c, err, 0) {
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
