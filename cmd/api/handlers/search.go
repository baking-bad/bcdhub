package handlers

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

type searchRequest struct {
	Text     string `form:"q"`
	Fields   string `form:"f,omitempty"`
	Networks string `form:"n,omitempty"`
	Offset   uint   `form:"o,omitempty"`
	DateFrom uint   `form:"s,omitempty"`
	DateTo   uint   `form:"e,omitempty"`
}

// Search -
func (ctx *Context) Search(c *gin.Context) {
	var req searchRequest
	if err := c.BindQuery(&req); err != nil {
		_ = c.AbortWithError(http.StatusBadRequest, err)
		return
	}

	var fields []string
	if req.Fields != "" {
		fields = strings.Split(req.Fields, ",")
	}
	var networks []string
	if req.Networks != "" {
		networks = strings.Split(req.Networks, ",")
	}
	contracts, err := ctx.ES.SearchByText(req.Text, int64(req.Offset), fields, networks, req.DateFrom, req.DateTo)
	if err != nil {
		_ = c.AbortWithError(http.StatusBadRequest, err)
		return
	}

	c.JSON(http.StatusOK, contracts)
}
