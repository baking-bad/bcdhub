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
// @Success 200 {object} TZIPResponse
// @Success 204 {object} gin.H
// @Failure 400 {object} Error
// @Failure 404 {object} Error
// @Failure 500 {object} Error
// @Router /v1/account/{network}/{address}/metadata [get]
func (ctx *Context) GetMetadata(c *gin.Context) {
	var req getContractRequest
	if err := c.BindUri(&req); ctx.handleError(c, err, http.StatusNotFound) {
		return
	}
	tzip, err := ctx.TZIP.Get(req.NetworkID(), req.Address)
	if err != nil {
		if ctx.Storage.IsRecordNotFound(err) {
			c.JSON(http.StatusNoContent, gin.H{})
		} else {
			ctx.handleError(c, err, 0)
		}
		return
	}

	if tzip.License.IsEmpty() {
		tzip.License = nil
	}

	var t TZIPResponse
	t.FromModel(tzip, true)
	c.JSON(http.StatusOK, t)
}
