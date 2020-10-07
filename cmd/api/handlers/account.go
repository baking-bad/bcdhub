package handlers

import (
	"net/http"

	"github.com/baking-bad/bcdhub/internal/elastic"
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
// @Router /account/{network}/{address} [get]
func (ctx *Context) GetInfo(c *gin.Context) {
	var req getContractRequest
	if err := c.BindUri(&req); handleError(c, err, http.StatusBadRequest) {
		return
	}

	stats, err := ctx.ES.GetOperationsStats(req.Network, req.Address)
	if handleError(c, err, 0) {
		return
	}

	rpc, err := ctx.GetRPC(req.Network)
	if handleError(c, err, 0) {
		return
	}

	balance, err := rpc.GetContractBalance(req.Address, 0)
	if handleError(c, err, 0) {
		return
	}

	alias := ctx.Aliases[req.Address]

	accountInfo := AccountInfo{
		Address:    req.Address,
		Network:    req.Network,
		Alias:      alias,
		TxCount:    stats.Count,
		Balance:    balance,
		LastAction: stats.LastAction,
		Tokens:     make([]TokenBalance, 0),
	}

	tokenBalances, err := ctx.ES.GetAccountBalances(req.Network, req.Address)
	if handleError(c, err, 0) {
		return
	}

	for key, b := range tokenBalances {
		tb := TokenBalance{
			Contract: key.Address,
			TokenID:  key.TokenID,
			Balance:  b,
		}
		token, err := ctx.ES.GetTokenMetadata(elastic.GetTokenMetadataContext{
			TokenID:  key.TokenID,
			Contract: key.Address,
			Network:  req.Network,
		})
		if err == nil {
			tb.Decimals = token[0].Decimals
			tb.Name = token[0].Name
			tb.Symbol = token[0].Symbol
		}

		accountInfo.Tokens = append(accountInfo.Tokens, tb)
	}

	c.JSON(http.StatusOK, accountInfo)
}
