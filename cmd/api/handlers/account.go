package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// GetInfo godoc
// @Summary Get account info
// @Description Get account info
// @Tags contract
// @ID get-account-info
// @Param network path string true "Network"
// @Param address path string true "Address" minlength(36) maxlength(36)
// @Accept  json
// @Produce  json
// @Success 200 {object} AccountInfo
// @Failure 400 {object} Error
// @Failure 500 {object} Error
// @Router /contract/{network}/{address}/info [get]
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

	c.JSON(http.StatusOK, AccountInfo{
		Address:    req.Address,
		Network:    req.Network,
		Alias:      alias,
		TxCount:    stats.Count,
		Balance:    balance,
		LastAction: stats.LastAction,
	})
}
