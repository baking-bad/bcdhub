package handlers

import (
	"net/http"
	"sort"

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

	tokenBalances, err := ctx.getAccountBalances(req.Network, req.Address)
	if ctx.handleError(c, err, 0) {
		return
	}
	accountInfo.Tokens = tokenBalances

	c.JSON(http.StatusOK, accountInfo)
}

func (ctx *Context) getAccountBalances(network, address string) ([]TokenBalance, error) {
	tokenBalances, err := ctx.TokenBalances.GetAccountBalances(network, address)
	if err != nil {
		return nil, err
	}

	result := make([]TokenBalance, 0)
	for _, balance := range tokenBalances {
		token, err := ctx.TokenMetadata.Get(tokenmetadata.GetContext{
			TokenID:  balance.TokenID,
			Contract: balance.Contract,
			Network:  network,
		})
		tb := TokenBalance{
			Balance: balance.Balance,
		}

		if err == nil && len(token) > 0 {
			tb.Decimals = token[0].Decimals
			tb.Name = token[0].Name
			tb.Symbol = token[0].Symbol
		}
		tb.Contract = balance.Contract
		tb.TokenID = balance.TokenID
		tb.Network = balance.Network

		result = append(result, tb)
	}

	sort.Slice(result, func(i, j int) bool {
		if result[i].Name == "" {
			return false
		} else if result[j].Name == "" {
			return true
		}

		return result[i].Name < result[j].Name
	})

	return result, nil
}
