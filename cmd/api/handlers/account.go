package handlers

import (
	"net/http"

	"github.com/baking-bad/bcdhub/internal/models/tokenmetadata"
	"github.com/gin-gonic/gin"
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
// @Failure 500 {object} Error
// @Router /v1/account/{network}/{address} [get]
func (ctx *Context) GetInfo(c *gin.Context) {
	var req getContractRequest
	if err := c.BindUri(&req); ctx.handleError(c, err, http.StatusBadRequest) {
		return
	}

	stats, err := ctx.Operations.GetStats(req.Network, req.Address)
	if ctx.handleError(c, err, 0) {
		return
	}
	block, err := ctx.Blocks.Last(req.Network)
	if ctx.handleError(c, err, 0) {
		return
	}

	rpc, err := ctx.GetRPC(req.Network)
	if ctx.handleError(c, err, 0) {
		return
	}
	balance, err := rpc.GetContractBalance(req.Address, block.Level)
	if ctx.handleError(c, err, 0) {
		return
	}

	accountInfo := AccountInfo{
		Address:    req.Address,
		Network:    req.Network,
		TxCount:    stats.Count,
		Balance:    balance,
		LastAction: stats.LastAction,
	}

	alias, err := ctx.TZIP.GetAlias(req.Network, req.Address)
	if err != nil {
		if !ctx.Storage.IsRecordNotFound(err) {
			ctx.handleError(c, err, 0)
			return
		}
	} else {
		accountInfo.Alias = alias.Name
	}

	c.JSON(http.StatusOK, accountInfo)
}

// GetAccountTokenBalances godoc
// @Summary Get account token balances
// @Description Get account token balances
// @Tags account
// @ID get-account-token-balances
// @Param network path string true "Network"
// @Param address path string true "Address" minlength(36) maxlength(36)
// @Param offset query integer false "Offset"
// @Param size query integer false "Requested count" minimum(0) maximum(10)
// @Param contract query string false "Contract address"
// @Accept  json
// @Produce  json
// @Success 200 {object} TokenBalances
// @Failure 400 {object} Error
// @Failure 500 {object} Error
// @Router /v1/account/{network}/{address}/token_balances [get]
func (ctx *Context) GetAccountTokenBalances(c *gin.Context) {
	var req getContractRequest
	if err := c.BindUri(&req); ctx.handleError(c, err, http.StatusBadRequest) {
		return
	}
	var queryParams tokenBalanceRequest
	if err := c.BindQuery(&queryParams); ctx.handleError(c, err, http.StatusBadRequest) {
		return
	}
	balances, err := ctx.getAccountBalances(req.Network, req.Address, queryParams)
	if ctx.handleError(c, err, 0) {
		return
	}
	c.JSON(http.StatusOK, balances)
}

func (ctx *Context) getAccountBalances(network, address string, req tokenBalanceRequest) (*TokenBalances, error) {
	tokenBalances, total, err := ctx.TokenBalances.GetAccountBalances(network, address, req.Contract, req.Size, req.Offset)
	if err != nil {
		return nil, err
	}

	response := TokenBalances{
		Balances: make([]TokenBalance, 0),
		Total:    total,
	}

	contextes := make([]tokenmetadata.GetContext, 0)
	balances := make(map[tokenmetadata.GetContext]string)

	for _, balance := range tokenBalances {
		c := tokenmetadata.GetContext{
			TokenID:  balance.TokenID,
			Contract: balance.Contract,
			Network:  network,
		}
		balances[c] = balance.Balance
		contextes = append(contextes, c)
	}

	tokens, err := ctx.TokenMetadata.GetAll(contextes...)
	if err != nil {
		return nil, err
	}

	for _, token := range tokens {
		c := tokenmetadata.GetContext{
			TokenID:  token.TokenID,
			Contract: token.Contract,
			Network:  network,
		}

		balance, ok := balances[c]
		if !ok {
			continue
		}

		delete(balances, c)

		tb := TokenBalance{
			Balance:       balance,
			TokenMetadata: TokenMetadataFromElasticModel(token, false),
		}

		response.Balances = append(response.Balances, tb)
	}

	for c, balance := range balances {
		response.Balances = append(response.Balances, TokenBalance{
			Balance: balance,
			TokenMetadata: TokenMetadata{
				Contract: c.Contract,
				TokenID:  c.TokenID,
				Network:  c.Network,
			},
		})
	}

	return &response, nil
}
