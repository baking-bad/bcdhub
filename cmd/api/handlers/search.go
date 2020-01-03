package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type searchRequest struct {
	Text string `form:"text"`
}

// Search -
func (ctx *Context) Search(c *gin.Context) {
	var req searchRequest
	if err := c.BindQuery(&req); err != nil {
		_ = c.AbortWithError(http.StatusBadRequest, err)
		return
	}

	contracts, err := ctx.ES.SearchByText(req.Text)
	if err != nil {
		_ = c.AbortWithError(http.StatusBadRequest, err)
		return
	}

	c.JSON(http.StatusOK, contracts)
}
