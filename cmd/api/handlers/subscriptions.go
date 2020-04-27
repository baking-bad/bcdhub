package handlers

import (
	"net/http"

	"github.com/baking-bad/bcdhub/internal/database"

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
	if len(subs) == 0 {
		return []Subscription{}, nil
	}

	ids := make([]string, len(subs))
	for i := range subs {
		ids[i] = subs[i].EntityID
	}

	contracts, err := ctx.ES.GetContractsByIDsWithSort(ids, "last_action", "desc")
	if err != nil {
		return nil, err
	}

	res := make([]Subscription, len(contracts))
	for i := range contracts {
		res[i] = Subscription{
			Contract: &Contract{
				Contract: &contracts[i],
			},
		}
		for j := range subs {
			if subs[j].EntityID == contracts[i].ID {
				res[i].SubscribedAt = subs[j].CreatedAt
				break
			}
		}
	}

	return res, nil
}
