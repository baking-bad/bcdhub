package handlers

import (
	"net/http"

	"github.com/baking-bad/bcdhub/internal/database"

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
	if handleError(c, err, 0) {
		return
	}

	c.JSON(http.StatusOK, PrepareSubscriptions(subscriptions))
}

// CreateSubscription -
func (ctx *Context) CreateSubscription(c *gin.Context) {
	var sub subRequest
	if err := c.ShouldBindJSON(&sub); handleError(c, err, http.StatusBadRequest) {
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
		Network:   sub.Network,
		WatchMask: buildWatchMask(sub),
	}

	if err := ctx.DB.UpsertSubscription(&subscription); handleError(c, err, 0) {
		return
	}

	c.JSON(http.StatusOK, subscription)
}

// DeleteSubscription -
func (ctx *Context) DeleteSubscription(c *gin.Context) {
	var sub subRequest
	if err := c.ShouldBindJSON(&sub); handleError(c, err, http.StatusBadRequest) {
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
		Network: sub.Network,
	}

	if err := ctx.DB.DeleteSubscription(&subscription); handleError(c, err, 0) {
		return
	}

	c.JSON(http.StatusOK, gin.H{})
}

// PrepareSubscription -
func PrepareSubscription(sub database.Subscription) (res Subscription) {
	res = buildSubFromWatchMask(sub.WatchMask)
	res.Address = sub.Address
	res.Network = sub.Network
	res.Alias = sub.Alias
	res.SubscribedAt = sub.CreatedAt
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

// Subscription flags
const (
	WatchSame uint = 1 << iota
	WatchSimilar
	WatchDeployed
	WatchMigrations
	WatchDeployments
	WatchCalls
	WatchErrors
)

func buildWatchMask(s subRequest) uint {
	var b uint

	if s.WatchSame {
		b = b | WatchSame
	}

	if s.WatchSimilar {
		b = b | WatchSimilar
	}

	if s.WatchDeployed {
		b = b | WatchDeployed
	}

	if s.WatchMigrations {
		b = b | WatchMigrations
	}

	if s.WatchDeployments {
		b = b | WatchDeployments
	}

	if s.WatchCalls {
		b = b | WatchCalls
	}

	if s.WatchErrors {
		b = b | WatchErrors
	}

	return b
}

func buildSubFromWatchMask(mask uint) Subscription {
	s := Subscription{}

	if mask&WatchSame != 0 {
		s.WatchSame = true
	}

	if mask&WatchSimilar != 0 {
		s.WatchSimilar = true
	}

	if mask&WatchDeployed != 0 {
		s.WatchDeployed = true
	}

	if mask&WatchMigrations != 0 {
		s.WatchMigrations = true
	}

	if mask&WatchDeployments != 0 {
		s.WatchDeployments = true
	}

	if mask&WatchCalls != 0 {
		s.WatchCalls = true
	}

	if mask&WatchErrors != 0 {
		s.WatchErrors = true
	}

	return s
}
