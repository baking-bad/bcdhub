package handlers

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

// AuthJWTRequired -
func (ctx *Context) AuthJWTRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		token := c.GetHeader("Authorization")

		id, err := ctx.OAUTH.GetIDFromToken(token)
		if err != nil {
			c.AbortWithError(http.StatusBadRequest, fmt.Errorf("invalid auth"))
			return
		}

		ctx.OAUTH.UserID = id

		c.Next()
	}

}
