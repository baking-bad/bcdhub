package handlers

import (
	"net/http"

	"github.com/aopoltorzhicky/bcdhub/internal/database"

	"github.com/gin-gonic/gin"
)

type subRequest struct {
	ID   string `json:"id"`
	Type string `json:"type"`
}

// ListSubscriptions -
func (ctx *Context) ListSubscriptions(c *gin.Context) {
	subscriptions, err := ctx.DB.ListSubscriptions(ctx.OAUTH.UserID)
	if err != nil {
		c.AbortWithError(http.StatusBadRequest, err)
		return
	}

	c.JSON(http.StatusOK, subscriptions)
}

// CreateSubscription -
func (ctx *Context) CreateSubscription(c *gin.Context) {
	var sub subRequest
	if err := c.ShouldBindJSON(&sub); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	subscription := database.Subscription{
		UserID:     ctx.OAUTH.UserID,
		EntityID:   sub.ID,
		EntityType: database.EntityType(sub.Type),
	}

	if err := ctx.DB.CreateSubscription(&subscription); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, subscription)
}

// DeleteSubscription -
func (ctx *Context) DeleteSubscription(c *gin.Context) {
	var sub subRequest
	if err := c.ShouldBindJSON(&sub); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	subscription := database.Subscription{
		UserID:     ctx.OAUTH.UserID,
		EntityID:   sub.ID,
		EntityType: database.EntityType(sub.Type),
	}

	if err := ctx.DB.DeleteSubscription(&subscription); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{})
}
