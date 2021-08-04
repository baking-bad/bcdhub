package handlers

import (
	"net/http"
	"strings"
	"time"

	"github.com/baking-bad/bcdhub/internal/bcd/forge"
	"github.com/baking-bad/bcdhub/internal/models"
	"github.com/baking-bad/bcdhub/internal/models/types"
	"github.com/baking-bad/bcdhub/internal/search"
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
// @Success 200 {object} search.Result
// @Failure 400 {object} Error
// @Failure 500 {object} Error
// @Router /v1/search [get]
func (ctx *Context) Search(c *gin.Context) {
	var req searchRequest
	if err := c.BindQuery(&req); ctx.handleError(c, err, http.StatusBadRequest) {
		return
	}

	var fields []string
	if req.Fields != "" {
		fields = strings.Split(req.Fields, ",")
	}
	filters := getSearchFilters(req)

	result, err := ctx.Searcher.ByText(req.Text, int64(req.Offset), fields, filters, req.Grouping != 0)
	if ctx.handleError(c, err, 0) {
		return
	}

	if result.Count == 0 {
		item := ctx.searchInMempool(req.Text)
		if item != nil {
			result.Items = append(result.Items, item)
			result.Count++
		}
	} else {
		ctx.searchPostprocessing(&result)
	}

	c.SecureJSON(http.StatusOK, result)
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
		indices := strings.Split(req.Indices, ",")
		arr := make([]string, 0)
		for i := range indices {
			if val, ok := indicesMap[indices[i]]; ok {
				arr = append(arr, val)
			}
		}
		filters["indices"] = arr
	}

	if req.Languages != "" {
		filters["languages"] = strings.Split(req.Languages, ",")
	}

	return filters
}

var indicesMap = map[string]string{
	"contract":       models.DocContracts,
	"operation":      models.DocOperations,
	"bigmapdiff":     models.DocBigMapDiff,
	"tzip":           models.DocTZIP,
	"token_metadata": models.DocTokenMetadata,
	"tezos_domain":   models.DocTezosDomains,
}

func (ctx *Context) searchInMempool(q string) *search.Item {
	if _, err := forge.UnforgeOpgHash(q); err != nil {
		return nil
	}

	if operation := ctx.getOperationFromMempool(q); operation != nil {
		return &search.Item{
			Type:  models.DocOperations,
			Value: operation.Hash,
			Body:  operation,
			Highlights: map[string][]string{
				"hash": {operation.Hash},
			},
		}
	}

	return nil
}

func (ctx *Context) searchPostprocessing(result *search.Result) {
	for i := range result.Items {
		switch typ := result.Items[i].Body.(type) {
		case *search.Contract:
			enity, err := ctx.Contracts.Get(types.NewNetwork(typ.Network), typ.Address)
			if err != nil {
				continue
			}

			ts := enity.LastAction.UTC()
			typ.TxCount = &enity.TxCount
			typ.LastAction = &ts
		case *search.Metadata:
			typ.Name = ctx.Sanitizer.Sanitize(typ.Name)
			typ.Description = ctx.Sanitizer.Sanitize(typ.Description)
			typ.Homepage = ctx.Sanitizer.Sanitize(typ.Homepage)
		case *search.Token:
			typ.Name = ctx.Sanitizer.Sanitize(typ.Name)
			typ.Symbol = ctx.Sanitizer.Sanitize(typ.Symbol)
		}
		for _, highlights := range result.Items[i].Highlights {
			for i := range highlights {
				highlights[i] = ctx.Sanitizer.Sanitize(highlights[i])
			}
		}
	}
}
