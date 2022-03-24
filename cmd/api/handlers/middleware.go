package handlers

import (
	"net/http"

	"github.com/baking-bad/bcdhub/internal/config"
	"github.com/baking-bad/bcdhub/internal/models/types"
	"github.com/gin-gonic/gin"
)

// NetworkMiddleware -
func NetworkMiddleware(ctxs config.Contexts) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req getByNetwork
		if err := c.BindUri(&req); err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, Error{Message: err.Error()})
			return
		}

		ctx, err := ctxs.Get(req.NetworkID())
		if err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, Error{Message: err.Error()})
			return
		}

		c.Set("context", ctx)

		c.Next()
	}
}

// MainnetMiddleware -
func MainnetMiddleware(ctxs config.Contexts) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, err := ctxs.Get(types.Mainnet)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, Error{Message: err.Error()})
			return
		}

		c.Set("context", ctx)

		c.Next()
	}
}

// ContextsMiddleware -
func ContextsMiddleware(ctxs config.Contexts) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set("contexts", ctxs)

		c.Next()
	}
}
