package handlers

import (
	"fmt"
	"net/http"

	"github.com/baking-bad/bcdhub/internal/helpers"

	"github.com/gin-gonic/gin"
)

// AuthJWTRequired -
func (ctx *Context) AuthJWTRequired() gin.HandlerFunc {
	if ctx.Config.API.Seed.Enabled {
		userID := ctx.OAUTH.UserID

		return func(c *gin.Context) {
			c.Set("userID", userID)
			c.Next()
		}
	}

	return func(c *gin.Context) {
		token := c.GetHeader("Authorization")

		id, err := ctx.OAUTH.GetIDFromToken(token)
		if handleError(c, err, http.StatusUnauthorized) {
			return
		}

		c.Set("userID", id)

		helpers.SetUserIDSentry(fmt.Sprintf("%v", id))

		c.Next()
	}

}

// IsAuthenticated -
func (ctx *Context) IsAuthenticated() gin.HandlerFunc {
	if ctx.Config.API.Seed.Enabled {
		userID := ctx.OAUTH.UserID

		return func(c *gin.Context) {
			c.Set("userID", userID)
			c.Next()
		}
	}

	return func(c *gin.Context) {
		token := c.GetHeader("Authorization")

		id, err := ctx.OAUTH.GetIDFromToken(token)
		if err != nil {
			return
		}

		c.Set("userID", id)

		c.Next()
	}

}
