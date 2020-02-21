package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// GetContractRating -
func (ctx *Context) GetContractRating(c *gin.Context) {
	var req getContractRequest
	if err := c.BindUri(&req); err != nil {
		_ = c.AbortWithError(http.StatusBadRequest, err)
		return
	}

	by := map[string]interface{}{
		"address": req.Address,
		"network": req.Network,
	}
	cntrID, err := ctx.ES.GetContractID(by)
	if err != nil {
		_ = c.AbortWithError(http.StatusBadRequest, err)
		return
	}

	rating, err := ctx.DB.GetSubscriptionRating(cntrID)
	if err != nil {
		c.AbortWithError(http.StatusBadRequest, err)
		return
	}

	c.JSON(http.StatusOK, rating)
}
