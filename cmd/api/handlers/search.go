package handlers

import (
	"net/http"
	"strings"
	"time"

	"github.com/baking-bad/bcdhub/internal/contractparser/stringer"
	"github.com/baking-bad/bcdhub/internal/elastic"
	"github.com/baking-bad/bcdhub/internal/models"
	"github.com/gin-gonic/gin"
)

// Search godoc
// @Summary Search in better-call
// @Description Search any data in contracts, operations and big map diff with filters
// @Tags search
// @ID search
// @Param q query string true "Query string"
// @Param f query string false "Comma-separated field names among which will search"
// @Param n query string false "Comma-separated networks list for searching"
// @Param o query integer false "Offset for pagination" mininum(0)
// @Param s query integer false "Return search result since given timestamp" mininum(0)
// @Param e query integer false "Return search result before given timestamp" mininum(0)
// @Param g query integer false "Grouping by contracts similarity. 0 - false, any others - true" Enums(0, 1)
// @Param i query string false "Comma-separated list of indices for searching. Values: contract, operation, bigmapdiff""
// @Param l query string false "Comma-separated list of languages for searching. Values: smartpy, liquidity, ligo, lorentz, michelson"
// @Accept  json
// @Produce  json
// @Success 200 {object} elastic.SearchResult
// @Failure 400 {object} Error
// @Failure 500 {object} Error
// @Router /search [get]
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
		filters["from"] = time.Unix(int64(req.DateFrom), 0).Format(time.RFC3339)
	}

	if req.DateTo > 0 {
		filters["to"] = time.Unix(int64(req.DateTo), 0).Format(time.RFC3339)
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
