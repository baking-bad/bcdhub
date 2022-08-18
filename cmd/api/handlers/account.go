package handlers

import (
	"net/http"

	"github.com/baking-bad/bcdhub/internal/config"
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
// @Failure 404 {object} Error
// @Failure 500 {object} Error
// @Router /v1/account/{network}/{address} [get]
func GetInfo() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.MustGet("context").(*config.Context)

		var req getAccountRequest
		if err := c.BindUri(&req); err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, Error{Message: err.Error()})
			return
		}

		acc, err := ctx.Accounts.Get(req.Address)
		if handleError(c, ctx.Storage, err, 0) {
			return
		}
		stats, err := ctx.Operations.ContractStats(acc.Address)
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
