package handlers

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// Search -
func (ctx *Context) Search(c *gin.Context) {
	var req searchRequest
	if err := c.BindQuery(&req); handleError(c, err, http.StatusBadRequest) {
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
	contracts, err := ctx.ES.SearchByText(req.Text, int64(req.Offset), fields, networks, req.DateFrom, req.DateTo, req.Grouping != 0)
	if handleError(c, err, 0) {
		return
	}

	c.JSON(http.StatusOK, contracts)
}
