package core

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/baking-bad/bcdhub/internal/logger"
	"github.com/baking-bad/bcdhub/internal/models"
	"github.com/baking-bad/bcdhub/internal/search"
	"github.com/pkg/errors"
)

const (
	defaultSize = 10
)

var sanitizeRegEx = regexp.MustCompile(`[\:]`)

func getFields(searchString string, filters map[string]interface{}, fields []string) ([]string, []string, Item, error) {
	var indices []string
	if val, ok := filters["indices"]; ok {
		indices = val.([]string)
		delete(filters, "indices")
	}

	scores, err := search.GetScores(searchString, fields, indices...)
	if err != nil {
		return nil, nil, nil, err
	}

	f := make([]string, 0)
	h := make(Item)
	for _, score := range scores.Scores {
		s := strings.Split(score, "^")
		h[s[0]] = Item{}
		f = append(f, s[0])
	}
	return f, scores.Indices, h, nil
}

func prepareSearchFilters(filters map[string]interface{}) (string, error) {
	builder := strings.Builder{}

	for k, v := range filters {
		if builder.Len() != 0 {
			builder.WriteString(" AND ")
		}
		switch k {
		case "from":
			val, ok := v.(string)
			if !ok {
				return "", errors.Errorf("Invalid type for 'from' filter (wait string): %T", v)
			}
			if val != "" {
				builder.WriteString(fmt.Sprintf("timestamp:{%s TO *}", val))
			}
		case "to":
			val, ok := v.(string)
			if !ok {
				return "", errors.Errorf("Invalid type for 'to' filter (wait string): %T", v)
			}
			if val != "" {
				builder.WriteString(fmt.Sprintf("timestamp:{* TO %s}", val))
			}
		case "networks":
			val, ok := v.([]string)
			if !ok {
				return "", errors.Errorf("Invalid type for 'network' filter (wait []string): %T", v)
			}
			if len(val) == 0 {
				continue
			}
			var str string
			if len(val) > 1 {
				str = fmt.Sprintf("network:(%s)", strings.Join(val, " OR "))
			} else {
				str = fmt.Sprintf("network:%s", val[0])
			}
			builder.WriteString(str)
		case "languages":
			val, ok := v.([]string)
			if !ok {
				return "", errors.Errorf("Invalid type for 'language' filter (wait []string): %T", v)
			}
			var str string
			if len(val) > 1 {
				str = fmt.Sprintf("language:(%s)", strings.Join(val, " OR "))
			} else {
				str = fmt.Sprintf("language:%s", val[0])
			}
			builder.WriteString(str)
		default:
			return "", errors.Errorf("Unknown search filter: %s", k)
		}
	}
	return builder.String(), nil
}

// SearchByText -
func (e *Elastic) SearchByText(text string, offset int64, fields []string, filters map[string]interface{}, group bool) (models.Result, error) {
	if text == "" {
		return models.Result{}, errors.Errorf("Empty search string. Please query something")
	}

	ctx, err := prepare(text, filters, fields)
	if err != nil {
		return models.Result{}, err
	}
	ctx.Offset = offset

	query := NewQuery().Query(
		QueryString(ctx.Text, ctx.Fields),
	)

	if group {
		query = grouping(ctx, query)
	}

	var response searchByTextResponse
	if err := e.Query(ctx.Indices, query, &response); err != nil {
		return models.Result{}, err
	}

	var items []models.Item
	if group {
		items, err = parseSearchGroupingResponse(response, offset)
	} else {
		items, err = parseSearchResponse(response)
	}
	if err != nil {
		logger.Error(err)
		return models.Result{}, nil
	}

	return models.Result{
		Items: items,
		Time:  response.Took,
		Count: response.Hits.Total.Value,
	}, nil
}

func parseSearchResponse(response searchByTextResponse) ([]models.Item, error) {
	items := make([]models.Item, 0)
	arr := response.Hits.Hits
	for i := range arr {
		val, err := search.Parse(arr[i].Index, arr[i].Highlight, arr[i].Source)
		if err != nil {
			return nil, err
		}
		if val == nil {
			continue
		}

		switch t := val.(type) {
		case models.Item:
			items = append(items, t)
		case []models.Item:
			items = append(items, t...)
		}
	}
	return items, nil
}

func parseSearchGroupingResponse(response searchByTextResponse, offset int64) ([]models.Item, error) {
	if len(response.Agg.Projects.Buckets) == 0 {
		return make([]models.Item, 0), nil
	}

	arr := response.Agg.Projects.Buckets
	lArr := int64(len(arr))
	items := make([]models.Item, 0)
	if offset > lArr {
		return items, nil
	}
	arr = arr[offset:]
	for i := range arr {
		searchItem := models.Item{}
		if arr[i].DocCount > 1 {
			searchItem.Group = models.NewGroup(arr[i].DocCount)
		}

		for j, item := range arr[i].Last.Hits.Hits {
			val, err := search.Parse(item.Index, item.Highlight, item.Source)
			if err != nil {
				return nil, err
			}
			if val == nil {
				continue
			}
			valItem := val.(models.Item)
			if j == 0 {
				searchItem.Type = valItem.Type
				searchItem.Body = valItem.Body
				searchItem.Value = valItem.Value
				searchItem.Highlights = item.Highlight
			} else {
				searchItem.Group.Top = append(searchItem.Group.Top, models.Top{
					Key:     valItem.Value,
					Network: valItem.Network,
				})
			}
		}
		items = append(items, searchItem)
	}
	return items, nil
}

func prepare(searchString string, filters map[string]interface{}, fields []string) (search.Context, error) {
	ctx := search.NewContext()

	if search.IsPtrSearch(searchString) {
		ctx.Text = strings.TrimPrefix(searchString, "ptr:")
		ctx.Indices = []string{models.DocBigMapDiff}
		ctx.Fields = []string{"ptr"}
	} else {
		internalFields, usingIndices, highlights, err := getFields(ctx.Text, filters, fields)
		if err != nil {
			return ctx, err
		}
		ctx.Indices = usingIndices
		ctx.Highlights = highlights
		ctx.Fields = internalFields
		ctx.Text = fmt.Sprintf("%s*", searchString)
	}

	filterString, err := prepareSearchFilters(filters)
	if err != nil {
		return ctx, err
	}
	if filterString != "" {
		ctx.Text = sanitizeRegEx.ReplaceAllString(ctx.Text, "\\${0}")
		ctx.Text = fmt.Sprintf("%s AND %s", filterString, ctx.Text)
	}
	return ctx, nil
}

func grouping(ctx search.Context, query Base) Base {
	topHits := Item{
		"top_hits": Item{
			"size": 1,
			"sort": List{
				Sort("_score", "desc"),
				Item{"last_action": Item{"order": "desc", "unmapped_type": "long"}},
				Sort("timestamp", "desc"),
			},
			"highlight": Item{
				"fields": ctx.Highlights,
			},
		},
	}

	query.Add(
		Aggs(
			AggItem{
				Name: "projects",
				Body: Item{
					"terms": Item{
						"script": `
							if (doc['_index'].value == "contract") {
								return doc['fingerprint.parameter'].value + '|' + doc['fingerprint.storage'].value + '|' + doc['fingerprint.code'].value
							} else if (doc['_index'].value == 'operation') {
								return doc['hash.keyword'].value
							} else if (doc['_index'].value == 'tzip') {
								return doc['network.keyword'].value + '|' + doc['address.keyword'].value
							} else if (doc['_index'].value == 'bigmapdiff') {
								return doc['key_hash.keyword'].value
							} else if (doc['_index'].value == 'tezos_domain') {
								return doc['name.keyword'].value + '|' + doc['network.keyword'].value
							} else if (doc['_index'].value == 'token_metadata') {
								return doc['network.keyword'].value + doc['contract.keyword'].value + doc['token_id'].value
							}`,
						"size": defaultSize + ctx.Offset,
						"order": List{
							Item{"bucket_score": "desc"},
							Item{"bucket_time": "desc"},
						},
					},
					"aggs": Item{
						"last": topHits,
						"bucket_score": Item{
							"max": Item{
								"script": "_score",
							},
						},
						"bucket_time": Item{
							"max": Item{
								"script": "if (doc.containsKey('last_action')) {return doc['last_action'].value} else {return doc['timestamp']}",
							},
						},
					},
				},
			},
		),
	).Zero()
	return query
}
