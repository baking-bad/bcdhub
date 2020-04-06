package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// GetTimeline -
func (ctx *Context) GetTimeline(c *gin.Context) {
	var req pageableRequest
	if err := c.BindQuery(&req); handleError(c, err, http.StatusBadRequest) {
		return
	}

	subscriptions, err := ctx.DB.ListSubscriptions(ctx.OAUTH.UserID)
	if handleError(c, err, 0) {
		return
	}

	contracts := make([]string, 0)

	for _, sub := range subscriptions {
		if sub.EntityType == "contract" {
			contracts = append(contracts, sub.EntityID)
		}
	}

	data, err := ctx.ES.GetTimeline(contracts, 20, req.Offset)
	if handleError(c, err, 0) {
		return
	}

	c.JSON(http.StatusOK, data)
}
