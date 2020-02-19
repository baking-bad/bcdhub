package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type userProfile struct {
	Login     string `json:"login"`
	AvatarURL string `json:"avatarURL"`
}

// GetUserProfile -
func (ctx *Context) GetUserProfile(c *gin.Context) {
	user, err := ctx.DB.GetUser(ctx.OAUTH.UserID)
	if err != nil {
		c.AbortWithError(http.StatusBadRequest, err)
		return
	}

	profile := userProfile{
		Login:     user.Login,
		AvatarURL: user.AvatarURL,
	}

	c.JSON(http.StatusOK, profile)
}
