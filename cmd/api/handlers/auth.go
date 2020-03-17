package handlers

import (
	"fmt"
	"net/http"

	"github.com/baking-bad/bcdhub/internal/helpers"

	"github.com/gin-gonic/gin"
)

// AuthJWTRequired -
func (ctx *Context) AuthJWTRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		token := c.GetHeader("Authorization")

		id, err := ctx.OAUTH.GetIDFromToken(token)
		if handleError(c, err, http.StatusUnauthorized) {
			return
		}

		ctx.OAUTH.UserID = id

		helpers.SetUserIDSentry(fmt.Sprintf("%v", id))

		c.Next()
	}

}

// IsAuthenticated -
func (ctx *Context) IsAuthenticated() gin.HandlerFunc {
	return func(c *gin.Context) {
		token := c.GetHeader("Authorization")

		id, err := ctx.OAUTH.GetIDFromToken(token)
		if err != nil {
			return
		}

		ctx.OAUTH.UserID = id

		c.Next()
	}

}
