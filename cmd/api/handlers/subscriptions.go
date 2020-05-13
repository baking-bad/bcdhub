package handlers

import (
	"net/http"

	"github.com/baking-bad/bcdhub/internal/database"

	"github.com/gin-gonic/gin"
)

// ListSubscriptions -
func (ctx *Context) ListSubscriptions(c *gin.Context) {
	subscriptions, err := ctx.DB.ListSubscriptions(CurrentUserID(c))
	if handleError(c, err, 0) {
		return
	}

	c.JSON(http.StatusOK, prepareSubscriptions(subscriptions))
}

// CreateSubscription -
func (ctx *Context) CreateSubscription(c *gin.Context) {
	var sub subRequest
	if err := c.ShouldBindJSON(&sub); handleError(c, err, http.StatusBadRequest) {
		return
	}

	subscription := database.Subscription{
		UserID:    CurrentUserID(c),
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

	subscription := database.Subscription{
		UserID:  CurrentUserID(c),
		Address: sub.Address,
		Network: sub.Network,
	}

	if err := ctx.DB.DeleteSubscription(&subscription); handleError(c, err, 0) {
		return
	}

	c.JSON(http.StatusOK, gin.H{})
}

func prepareSubscriptions(subs []database.Subscription) []Subscription {
	res := make([]Subscription, len(subs))

	for i, sub := range subs {
		res[i] = buildSubFromWatchMask(sub.WatchMask)
		res[i].Address = sub.Address
		res[i].Network = sub.Network
		res[i].Alias = sub.Alias
		res[i].SubscribedAt = sub.CreatedAt
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

func set(b, flag uint) uint { return b | flag }
func has(b, flag uint) bool { return b&flag != 0 }

func buildWatchMask(s subRequest) uint {
	var b uint

	if s.WatchSame {
		b = set(b, WatchSame)
	}

	if s.WatchSimilar {
		b = set(b, WatchSimilar)
	}

	if s.WatchDeployed {
		b = set(b, WatchDeployed)
	}

	if s.WatchMigrations {
		b = set(b, WatchMigrations)
	}

	if s.WatchDeployments {
		b = set(b, WatchDeployments)
	}

	if s.WatchCalls {
		b = set(b, WatchCalls)
	}

	if s.WatchErrors {
		b = set(b, WatchErrors)
	}

	return b
}

func buildSubFromWatchMask(mask uint) Subscription {
	s := Subscription{}

	if has(mask, WatchSame) {
		s.WatchSame = true
	}

	if has(mask, WatchSimilar) {
		s.WatchSimilar = true
	}

	if has(mask, WatchDeployed) {
		s.WatchDeployed = true
	}

	if has(mask, WatchMigrations) {
		s.WatchMigrations = true
	}

	if has(mask, WatchDeployments) {
		s.WatchDeployments = true
	}

	if has(mask, WatchCalls) {
		s.WatchCalls = true
	}

<<<<<<< HEAD
	res := make([]Subscription, len(contracts))
	for i := range contracts {
		var contract Contract
		contract.FromModel(contracts[i])
		res[i] = Subscription{
			Contract: &contract,
		}
		for j := range subs {
			if subs[j].EntityID == contracts[i].ID {
				res[i].SubscribedAt = subs[j].CreatedAt
				break
			}
		}
=======
	if has(mask, WatchErrors) {
		s.WatchErrors = true
>>>>>>> Refactor subscriptions. Refactor database package. Fix auth. Seed data for sandbox
	}

	return s
}
