package handlers

import (
	"net/http"

	"github.com/baking-bad/bcdhub/internal/elastic"
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
// @Success 200 {object} models.TZIP
// @Success 204 {object} gin.H
// @Failure 400 {object} Error
// @Failure 500 {object} Error
// @Router /account/{network}/{address}/metadata [get]
func (ctx *Context) GetMetadata(c *gin.Context) {
	var req getContractRequest
	if err := c.BindUri(&req); handleError(c, err, http.StatusBadRequest) {
		return
	}
	tzip, err := ctx.ES.GetTZIP(req.Network, req.Address)
	if err != nil {
		if elastic.IsRecordNotFound(err) {
			c.JSON(http.StatusNoContent, gin.H{})
		} else {
			handleError(c, err, 0)
		}
		return
	}
	c.JSON(http.StatusOK, tzip)
}
