package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type userProfile struct {
	Login         string         `json:"login"`
	AvatarURL     string         `json:"avatarURL"`
	Subscriptions []Subscription `json:"subscriptions"`
}

// GetUserProfile -
func (ctx *Context) GetUserProfile(c *gin.Context) {
	user, err := ctx.DB.GetUser(ctx.OAUTH.UserID)
	if err != nil {
		c.AbortWithError(http.StatusBadRequest, err)
		return
	}

	subscriptions, err := ctx.DB.ListSubscriptionsWithLimit(ctx.OAUTH.UserID, 10)
	if err != nil {
		c.AbortWithError(http.StatusBadRequest, err)
		return
	}

	subs, err := ctx.prepareSubscription(subscriptions)
	if err != nil {
		c.AbortWithError(http.StatusBadRequest, err)
		return
	}

	profile := userProfile{
		Login:         user.Login,
		AvatarURL:     user.AvatarURL,
		Subscriptions: subs,
	}

	c.JSON(http.StatusOK, profile)
}
