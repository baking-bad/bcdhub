package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// GetUserProfile -
func (ctx *Context) GetUserProfile(c *gin.Context) {
	user, err := ctx.DB.GetUser(ctx.OAUTH.UserID)
	if handleError(c, err, 0) {
		return
	}

	subscriptions, err := ctx.DB.ListSubscriptionsWithLimit(ctx.OAUTH.UserID, 10)
	if handleError(c, err, 0) {
		return
	}

	subs, err := ctx.prepareSubscription(subscriptions)
	if handleError(c, err, 0) {
		return
	}

	profile := userProfile{
		Login:         user.Login,
		AvatarURL:     user.AvatarURL,
		Subscriptions: subs,
	}

	c.JSON(http.StatusOK, profile)
}
