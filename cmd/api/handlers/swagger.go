package handlers

import (
	"github.com/baking-bad/bcdhub/internal/config"
	"github.com/gin-gonic/gin"
	"github.com/swaggo/swag"
)

// GetSwaggerDoc -
func GetSwaggerDoc() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.MustGet("context").(*config.Context)

		doc, err := swag.ReadDoc()
		if handleError(c, ctx.Storage, err, 0) {
			return
		}

		c.Header("Content-Type", "application/json")
		_, err = c.Writer.Write([]byte(doc))
		if handleError(c, ctx.Storage, err, 0) {
			return
		}
	}
}
