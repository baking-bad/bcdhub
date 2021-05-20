package handlers

import (
	"fmt"
	"net/http"

	"github.com/baking-bad/bcdhub/internal/database"
	"github.com/baking-bad/bcdhub/internal/models/types"

	"github.com/gin-gonic/gin"
)

// ListSubscriptions -
func (ctx *Context) ListSubscriptions(c *gin.Context) {
	userID := CurrentUserID(c)
	if userID == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user"})
		return
	}

	subscriptions, err := ctx.DB.ListSubscriptions(userID)
	if ctx.handleError(c, err, 0) {
		return
	}

	c.JSON(http.StatusOK, PrepareSubscriptions(subscriptions))
}

// CreateSubscription -
func (ctx *Context) CreateSubscription(c *gin.Context) {
	var sub subRequest
	if err := c.ShouldBindJSON(&sub); ctx.handleError(c, err, http.StatusBadRequest) {
		return
	}
	if sub.SentryEnabled && sub.SentryDSN == "" {
		ctx.handleError(c, fmt.Errorf("You have to set `Sentry DSN` when sentry notifications is enabled"), http.StatusBadRequest)
		return
	}

	userID := CurrentUserID(c)
	if userID == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user"})
		return
	}

	subscription := database.Subscription{
		UserID:    userID,
		Address:   sub.Address,
		Network:   types.NewNetwork(sub.Network),
		Alias:     sub.Alias,
		WatchMask: sub.getMask(),
		SentryDSN: sub.SentryDSN,
	}

	if err := ctx.DB.UpsertSubscription(&subscription); ctx.handleError(c, err, 0) {
		return
	}

	c.JSON(http.StatusOK, subscription)
}

// DeleteSubscription -
func (ctx *Context) DeleteSubscription(c *gin.Context) {
	var sub subRequest
	if err := c.ShouldBindJSON(&sub); ctx.handleError(c, err, http.StatusBadRequest) {
		return
	}

	userID := CurrentUserID(c)
	if userID == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user"})
		return
	}

	subscription := database.Subscription{
		UserID:  userID,
		Address: sub.Address,
		Network: types.NewNetwork(sub.Network),
	}

	if err := ctx.DB.DeleteSubscription(&subscription); ctx.handleError(c, err, 0) {
		return
	}

	c.JSON(http.StatusOK, gin.H{})
}

// PrepareSubscription -
func PrepareSubscription(sub database.Subscription) (res Subscription) {
	res = newSubscriptionWithMask(sub.WatchMask)
	res.Address = sub.Address
	res.Network = sub.Network.String()
	res.Alias = sub.Alias
	res.SubscribedAt = sub.CreatedAt
	res.SentryDSN = sub.SentryDSN
	return
}

// PrepareSubscriptions -
func PrepareSubscriptions(subs []database.Subscription) []Subscription {
	res := make([]Subscription, len(subs))

	for i, sub := range subs {
		res[i] = PrepareSubscription(sub)
	}

	return res
}
