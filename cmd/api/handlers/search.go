package handlers

import (
	"net/http"
	"strings"

	"github.com/baking-bad/bcdhub/internal/contractparser/stringer"
	"github.com/baking-bad/bcdhub/internal/elastic"
	"github.com/baking-bad/bcdhub/internal/models"
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

	result, err := ctx.ES.SearchByText(req.Text, int64(req.Offset), fields, filters, req.Grouping != 0)
	if handleError(c, err, 0) {
		return
	}
	result, err = postProcessing(result)
	if handleError(c, err, 0) {
		return
	}

	c.JSON(http.StatusOK, result)
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

func postProcessing(result elastic.SearchResult) (elastic.SearchResult, error) {
	for i := range result.Items {
		switch result.Items[i].Type {
		case elastic.DocBigMapDiff:
			bmd := result.Items[i].Body.(models.BigMapDiff)
			key, err := stringer.StringifyInterface(bmd.Key)
			if err != nil {
				return result, err
			}

			result.Items[i].Body = SearchBigMapDiff{
				Ptr:       bmd.Ptr,
				Key:       key,
				KeyHash:   bmd.KeyHash,
				Value:     bmd.Value,
				Level:     bmd.Level,
				Address:   bmd.Address,
				Network:   bmd.Network,
				Timestamp: bmd.Timestamp,
				FoundBy:   bmd.FoundBy,
			}
		}
	}
	return result, nil
}
