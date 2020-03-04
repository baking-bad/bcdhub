package handlers

import (
	"net/http"

	"github.com/aopoltorzhicky/bcdhub/internal/database"

	"github.com/gin-gonic/gin"
)

// ListSubscriptions -
func (ctx *Context) ListSubscriptions(c *gin.Context) {
	subscriptions, err := ctx.DB.ListSubscriptions(ctx.OAUTH.UserID)
	if handleError(c, err, 0) {
		return
	}

	res, err := ctx.prepareSubscription(subscriptions)
	if handleError(c, err, 0) {
		return
	}

	c.JSON(http.StatusOK, res)
}

// CreateSubscription -
func (ctx *Context) CreateSubscription(c *gin.Context) {
	var sub subRequest
	if err := c.ShouldBindJSON(&sub); handleError(c, err, http.StatusBadRequest) {
		return
	}

	subscription := database.Subscription{
		UserID:     ctx.OAUTH.UserID,
		EntityID:   sub.ID,
		EntityType: database.EntityType(sub.Type),
	}

	if err := ctx.DB.CreateSubscription(&subscription); handleError(c, err, 0) {
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
		UserID:     ctx.OAUTH.UserID,
		EntityID:   sub.ID,
		EntityType: database.EntityType(sub.Type),
	}

	if err := ctx.DB.DeleteSubscription(&subscription); handleError(c, err, 0) {
		return
	}

	c.JSON(http.StatusOK, gin.H{})
}

func (ctx *Context) prepareSubscription(subs []database.Subscription) ([]Subscription, error) {
	res := make([]Subscription, len(subs))
	for i, s := range subs {
		c, err := ctx.ES.GetContractByID(s.EntityID)
		if err != nil {
			return nil, err
		}
		contract := Contract{
			Contract: &c,
		}
		if err := ctx.setAlias(&contract); err != nil {
			return nil, err
		}

		res[i] = Subscription{
			Contract:     &contract,
			SubscribedAt: s.CreatedAt,
		}
	}
	return res, nil
}
