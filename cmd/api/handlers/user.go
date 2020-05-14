package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// GetUserProfile -
func (ctx *Context) GetUserProfile(c *gin.Context) {
	userID := CurrentUserID(c)
	if userID == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user"})
		return
	}

	user, err := ctx.DB.GetUser(userID)
	if handleError(c, err, 0) {
		return
	}

	subscriptions, err := ctx.DB.ListSubscriptionsWithLimit(userID, 10)
	if handleError(c, err, 0) {
		return
	}

	profile := userProfile{
		Login:         user.Login,
		AvatarURL:     user.AvatarURL,
		Subscriptions: prepareSubscriptions(subscriptions),
		MarkReadAt:    user.MarkReadAt,
	}

	c.JSON(http.StatusOK, profile)
}
