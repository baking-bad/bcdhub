package handlers

import (
	"net/http"

	"github.com/baking-bad/bcdhub/internal/database"
	"github.com/baking-bad/bcdhub/internal/elastic"
	"github.com/gin-gonic/gin"
)

// GetEvents -
func (ctx *Context) GetEvents(c *gin.Context) {
	userID := CurrentUserID(c)
	if userID == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user"})
		return
	}

	var pageReq pageableRequest
	if err := c.BindQuery(&pageReq); handleError(c, err, http.StatusBadRequest) {
		return
	}

	subscriptions, err := ctx.DB.ListSubscriptions(userID)
	if handleError(c, err, 0) {
		return
	}

	events, err := ctx.getEvents(subscriptions, pageReq.Size, pageReq.Offset)
	if handleError(c, err, 0) {
		return
	}

	c.JSON(http.StatusOK, events)
}

func (ctx *Context) getEvents(subscriptions []database.Subscription, size, offset int64) ([]elastic.Event, error) {
	subs := make([]elastic.SubscriptionRequest, len(subscriptions))
	for i := range subscriptions {
		contract, err := ctx.ES.GetContract(map[string]interface{}{
			"address": subscriptions[i].Address,
			"network": subscriptions[i].Network,
		})
		if err != nil {
			return []elastic.Event{}, err
		}

		subs[i] = elastic.SubscriptionRequest{
			Address:   subscriptions[i].Address,
			Network:   subscriptions[i].Network,
			Alias:     subscriptions[i].Alias,
			Hash:      contract.Hash,
			ProjectID: contract.ProjectID,

			WithSame:        subscriptions[i].WatchMask&WatchSame != 0,
			WithSimilar:     subscriptions[i].WatchMask&WatchSimilar != 0,
			WithDeployed:    subscriptions[i].WatchMask&WatchDeployed != 0,
			WithMigrations:  subscriptions[i].WatchMask&WatchMigrations != 0,
			WithDeployments: subscriptions[i].WatchMask&WatchDeployments != 0,
			WithCalls:       subscriptions[i].WatchMask&WatchCalls != 0,
			WithErrors:      subscriptions[i].WatchMask&WatchErrors != 0,
		}
	}

	return ctx.ES.GetEvents(subs, size, offset)
}
