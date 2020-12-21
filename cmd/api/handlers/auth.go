package handlers

import (
	"fmt"
	"net/http"

	"github.com/baking-bad/bcdhub/internal/helpers"

	"github.com/gin-gonic/gin"
)

// AuthJWTRequired -
func (ctx *Context) AuthJWTRequired() gin.HandlerFunc {
	if ctx.Config.API.SeedEnabled {
		return ctx.skipAuth()
	}

	return func(c *gin.Context) {
		userID, err := ctx.getUserFromToken(c)
		if ctx.handleError(c, err, http.StatusUnauthorized) {
			return
		}

		helpers.SetUserIDSentry(fmt.Sprintf("%v", userID))
		setUserIDAndNext(userID, c)
	}

}

// IsAuthenticated -
func (ctx *Context) IsAuthenticated() gin.HandlerFunc {
	if ctx.Config.API.SeedEnabled {
		return ctx.skipAuth()
	}

	return func(c *gin.Context) {
		userID, err := ctx.getUserFromToken(c)
		if err != nil {
			return
		}

		setUserIDAndNext(userID, c)
	}

}

func (ctx *Context) getUserFromToken(c *gin.Context) (uint, error) {
	token := c.GetHeader("Authorization")
	return ctx.OAUTH.GetIDFromToken(token)
}

func (ctx *Context) skipAuth() func(*gin.Context) {
	userID := ctx.OAUTH.UserID

	return func(c *gin.Context) {
		setUserIDAndNext(userID, c)
	}
}

func setUserIDAndNext(userID uint, c *gin.Context) {
	c.Set("userID", userID)
	c.Next()
}
