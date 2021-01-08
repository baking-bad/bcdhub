package handlers

import (
	"github.com/gin-gonic/gin"
	"github.com/swaggo/swag"
)

// GetSwaggerDoc -
func (ctx *Context) GetSwaggerDoc(c *gin.Context) {
	doc, err := swag.ReadDoc()
	if ctx.handleError(c, err, 0) {
		return
	}

	c.Header("Content-Type", "application/json")
	_, err = c.Writer.Write([]byte(doc))
	if ctx.handleError(c, err, 0) {
		return
	}
}
