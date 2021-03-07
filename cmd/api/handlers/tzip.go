package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// GetMetadata godoc
// @Summary Get metadata for account
// @Description Returns full metadata for account
// @Tags account
// @ID get-account-tzip
// @Param network path string true "Network"
// @Param address path string true "KT or tz address" minlength(36) maxlength(36)
// @Accept  json
// @Produce  json
// @Success 200 {object} tzip.TZIP
// @Success 204 {object} gin.H
// @Failure 400 {object} Error
// @Failure 500 {object} Error
// @Router /v1/account/{network}/{address}/metadata [get]
func (ctx *Context) GetMetadata(c *gin.Context) {
	var req getContractRequest
	if err := c.BindUri(&req); ctx.handleError(c, err, http.StatusBadRequest) {
		return
	}
	tzip, err := ctx.TZIP.Get(req.Network, req.Address)
	if err != nil {
		if ctx.Storage.IsRecordNotFound(err) {
			c.JSON(http.StatusNoContent, gin.H{})
		} else {
			ctx.handleError(c, err, 0)
		}
		return
	}

	c.JSON(http.StatusOK, TZIPResponse{
		TZIP16:  tzip.TZIP16,
		TZIP20:  tzip.TZIP20,
		Domain:  tzip.Domain,
		Extras:  tzip.Extras,
		Address: tzip.Address,
		Network: tzip.Network,
		Name:    tzip.Name,
	})
}
