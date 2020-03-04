package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// GetContractRating -
func (ctx *Context) GetContractRating(c *gin.Context) {
	var req getContractRequest
	if err := c.BindUri(&req); handleError(c, err, http.StatusBadRequest) {
		return
	}

	by := map[string]interface{}{
		"address": req.Address,
		"network": req.Network,
	}
	cntrID, err := ctx.ES.GetContractID(by)
	if handleError(c, err, 0) {
		return
	}

	rating, err := ctx.DB.GetSubscriptionRating(cntrID)
	if handleError(c, err, 0) {
		return
	}

	c.JSON(http.StatusOK, rating)
}
