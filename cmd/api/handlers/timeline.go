package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// GetTimeline -
func (ctx *Context) GetTimeline(c *gin.Context) {
	subscriptions, err := ctx.DB.ListSubscriptions(ctx.OAUTH.UserID)
	if handleError(c, err, 0) {
		return
	}

	projectIDs := make([]string, 0)
	contracts := make([]string, 0)

	for _, sub := range subscriptions {
		if sub.EntityType == "contract" {
			contracts = append(contracts, sub.EntityID)
		} else if sub.EntityType == "project" {
			projectIDs = append(projectIDs, sub.EntityID)
		}
	}

	data, err := ctx.ES.GetTimeline(projectIDs, contracts)
	if handleError(c, err, 0) {
		return
	}

	c.JSON(http.StatusOK, data)
}
