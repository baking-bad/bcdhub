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
	filters := getSearchFilters(req)

	contracts, err := ctx.ES.SearchByText(req.Text, int64(req.Offset), fields, filters, req.Grouping != 0)
	if handleError(c, err, 0) {
		return
	}

	c.JSON(http.StatusOK, contracts)
}

func getSearchFilters(req searchRequest) map[string]interface{} {
	filters := map[string]interface{}{}

	if req.DateFrom > 0 {
		filters["from"] = req.DateFrom
	}

	if req.DateTo > 0 {
		filters["to"] = req.DateTo
	}

	if req.Networks != "" {
		filters["networks"] = strings.Split(req.Networks, ",")
	}

	if req.Indices != "" {
		filters["indices"] = strings.Split(req.Indices, ",")
	}

	if req.Languages != "" {
		filters["languages"] = strings.Split(req.Languages, ",")
	}

	return filters
}
