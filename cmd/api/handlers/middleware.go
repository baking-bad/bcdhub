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
		if err := c.ShouldBindUri(&req); err != nil {
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

// MaxRequestBodySize - upper bound for request bodies. Contract scripts are
// limited by the protocol to tens of kilobytes, so this leaves ample room for
// their Micheline JSON while capping the input to the decoders, whose work
// grows faster than linearly on adversarial input.
const MaxRequestBodySize = 1 << 20

// BodyLimitMiddleware - rejects request bodies larger than limit bytes
func BodyLimitMiddleware(limit int64) gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.ContentLength > limit {
			c.AbortWithStatusJSON(http.StatusRequestEntityTooLarge, Error{Message: "request body too large"})
			return
		}
		if c.Request.Body != nil {
			c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, limit)
		}
		c.Next()
	}
}
