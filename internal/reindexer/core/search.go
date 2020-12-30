package core

import (
	"fmt"
	"strings"
	"time"

	"github.com/baking-bad/bcdhub/internal/models"
	"github.com/baking-bad/bcdhub/internal/models/bigmapdiff"
	"github.com/baking-bad/bcdhub/internal/models/contract"
	"github.com/baking-bad/bcdhub/internal/models/operation"
	"github.com/baking-bad/bcdhub/internal/models/tezosdomain"
	"github.com/baking-bad/bcdhub/internal/models/tzip"
	"github.com/baking-bad/bcdhub/internal/search"
	"github.com/pkg/errors"
	"github.com/restream/reindexer"
)

// SearchByText -
func (r *Reindexer) SearchByText(text string, offset int64, fields []string, filters map[string]interface{}, group bool) (models.Result, error) {
	result := models.Result{}
	if text == "" {
		return result, errors.Errorf("Empty search string. Please query something")
	}

	query, err := r.prepareSearchQuery(text, filters, fields, offset)
	if err != nil {
		return result, err
	}

	start := time.Now()
	it := query.Exec()
	defer it.Close()

	if it.Error() != nil {
		return result, err
	}
	items, err := parseSearchResponse(it)
	if err != nil {
		return result, err
	}
	result.Time = time.Since(start).Milliseconds()
	result.Count = int64(it.TotalCount())
	result.Items = items

	return result, nil
}

func (r *Reindexer) prepareSearchQuery(searchString string, filters map[string]interface{}, fields []string, offset int64) (*reindexer.Query, error) {
	ctx := search.NewContext()

	if search.IsPtrSearch(searchString) {
		ctx.Text = strings.TrimPrefix(searchString, "ptr:")
		ctx.Indices = []string{models.DocBigMapDiff}
		ctx.Fields = []string{"ptr"}
	} else {
		info, err := getFields(ctx.Text, filters, fields)
		if err != nil {
			return nil, err
		}
		ctx.Indices = info.Indices
		ctx.Fields = info.Scores
		ctx.Text = fmt.Sprintf("%s*", searchString)
	}
	ctx.Offset = offset

	return r.buildSeacrhQuery(ctx, filters)
}

func (r *Reindexer) buildSeacrhQuery(ctx search.Context, filters map[string]interface{}) (*reindexer.Query, error) {
	var query *reindexer.Query
	for i := range ctx.Indices {
		subQuery := r.Query(ctx.Indices[i])
		for _, field := range ctx.Fields {
			subQuery = subQuery.Match(field, ctx.Text)
		}
		if err := prepareFilters(filters, subQuery); err != nil {
			return nil, err
		}

		subQuery = subQuery.Offset(int(ctx.Offset)).Functions("text.highlight(<em>,</em>)")

		if query == nil {
			query = subQuery
		} else {
			query.Merge(subQuery)
		}
	}

	query = query.ReqTotal()

	return query, nil
}

func getFields(searchString string, filters map[string]interface{}, fields []string) (search.ScoreInfo, error) {
	var indices []string
	if val, ok := filters["indices"]; ok {
		indices = val.([]string)
		delete(filters, "indices")
	}

	return search.GetScores(searchString, fields, indices...)
}

func prepareFilters(filters map[string]interface{}, query *reindexer.Query) error {
	for field, value := range filters {
		switch field {
		case "from":
			query = query.Where("timestamp", reindexer.GT, value)
		case "to":
			query = query.Where("timestamp", reindexer.LT, value)
		case "networks":
			networks, ok := value.([]string)
			if !ok {
				return errors.Errorf("Invalid type for 'network' filter (wait []string): %T", value)
			}
			query = query.Match("network", networks...)
		case "languages":
			languages, ok := value.([]string)
			if !ok {
				return errors.Errorf("Invalid type for 'network' filter (wait []string): %T", value)
			}
			query = query.Match("language", languages...)
		default:
			return errors.Errorf("Unknown search filter: %s", field)
		}
	}
	return nil
}

func parseSearchResponse(it *reindexer.Iterator) ([]models.Item, error) {
	items := make([]models.Item, 0)
	for it.Next() {
		searchItem := models.Item{}

		switch elem := it.Object().(type) {
		case contract.Contract:
			searchItem.Type = models.DocContracts
			searchItem.Value = elem.Address
			searchItem.Body = elem
			searchItem.Network = elem.Network
		case operation.Operation:
			searchItem.Type = models.DocOperations
			searchItem.Value = elem.Hash
			searchItem.Body = elem
			searchItem.Network = elem.Network
		case bigmapdiff.BigMapDiff:
			searchItem.Type = models.DocBigMapDiff
			searchItem.Value = elem.KeyHash
			searchItem.Body = elem
			searchItem.Network = elem.Network
		case tezosdomain.TezosDomain:
			searchItem.Type = models.DocTezosDomains
			searchItem.Value = elem.Address
			searchItem.Body = elem
			searchItem.Network = elem.Network
		case tzip.TZIP:
			searchItem.Value = elem.Address
			searchItem.Network = elem.Network
			searchItem.Type = search.MetadataSearchType
		default:
			return nil, errors.Errorf("Unknown search type")
		}

		items = append(items, searchItem)
	}
	return items, nil
}
