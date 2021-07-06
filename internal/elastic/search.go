package elastic

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

var wordRegexp = regexp.MustCompile(`^\w*$`)

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

	h := make(Item)
	for _, score := range scores.Scores {
		s := strings.Split(score, "^")
		h[s[0]] = Item{}
	}
	return scores.Scores, scores.Indices, h, nil
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

// ByText -
func (e *Elastic) ByText(text string, offset int64, fields []string, filters map[string]interface{}, group bool) (search.Result, error) {
	if text == "" {
		return search.Result{}, errors.Errorf("Empty search string. Please query something")
	}

	ctx, err := prepare(text, filters, fields)
	if err != nil {
		return search.Result{}, err
	}
	ctx.Offset = offset

	query := NewQuery().Query(
		QueryString(ctx.Text, ctx.Fields),
	)

	if group {
		query = grouping(ctx, query)
	} else {
		query.From(offset)
	}

	var response searchByTextResponse
	if err := e.query(ctx.Indices, query, &response); err != nil {
		return search.Result{}, err
	}

	var items []*search.Item
	if group {
		items, err = parseSearchGroupingResponse(response, offset)
	} else {
		items, err = parseSearchResponse(response)
	}
	if err != nil {
		logger.Err(err)
		return search.Result{}, nil
	}

	return search.Result{
		Items: items,
		Time:  response.Took,
		Count: response.Hits.Total.Value,
	}, nil
}

func parseSearchResponse(response searchByTextResponse) ([]*search.Item, error) {
	items := make([]*search.Item, 0)
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
		case *search.Item:
			items = append(items, t)
		case []*search.Item:
			items = append(items, t...)
		}
	}
	return items, nil
}

func parseSearchGroupingResponse(response searchByTextResponse, offset int64) ([]*search.Item, error) {
	if len(response.Agg.Projects.Buckets) == 0 {
		return make([]*search.Item, 0), nil
	}

	arr := response.Agg.Projects.Buckets
	lArr := int64(len(arr))
	items := make([]*search.Item, 0)
	if offset > lArr {
		return items, nil
	}
	arr = arr[offset:]
	for i := range arr {
		searchItem := &search.Item{}
		if arr[i].DocCount > 1 {
			searchItem.Group = search.NewGroup(arr[i].DocCount)
		}

		for j, item := range arr[i].Last.Hits.Hits {
			val, err := search.Parse(item.Index, item.Highlight, item.Source)
			if err != nil {
				return nil, err
			}
			if val == nil {
				continue
			}
			valItem := val.(*search.Item)
			if j == 0 {
				searchItem.Type = typeMap[valItem.Type]
				searchItem.Body = valItem.Body
				searchItem.Value = valItem.Value
				searchItem.Highlights = item.Highlight
			} else {
				searchItem.Group.Top = append(searchItem.Group.Top, search.Top{
					Key:     valItem.Value,
					Network: valItem.Network,
				})
			}
		}
		items = append(items, searchItem)
	}
	return items, nil
}

var typeMap = map[string]string{
	models.DocContracts:     "contract",
	models.DocOperations:    "operation",
	models.DocBigMapDiff:    "bigmapdiff",
	models.DocTokenMetadata: "token_metadata",
	models.DocTezosDomains:  "tezos_domain",
	models.DocTZIP:          "tzip",
}

func prepare(searchString string, filters map[string]interface{}, fields []string) (search.Context, error) {
	ctx := search.NewContext()

	if search.IsPtrSearch(searchString) {
		ctx.Text = strings.TrimPrefix(searchString, "ptr:")
		ctx.Indices = []string{models.DocBigMapDiff}
		ctx.Fields = []string{"ptr"}
	} else {
		internalFields, usingIndices, highlights, err := getFields(searchString, filters, fields)
		if err != nil {
			return ctx, err
		}
		ctx.Indices = usingIndices
		ctx.Highlights = highlights
		ctx.Fields = internalFields
		if wordRegexp.MatchString(searchString) {
			ctx.Text = fmt.Sprintf("%s*", searchString)
		} else {
			ctx.Text = fmt.Sprintf("\"%s*\"", searchString)
		}
	}

	filterString, err := prepareSearchFilters(filters)
	if err != nil {
		return ctx, err
	}
	if filterString != "" {
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
							if (doc['_index'].value == "contracts") {
								return doc['hash.keyword'].value
							} else if (doc['_index'].value == 'operations') {
								return doc['hash.keyword'].value
							} else if (doc['_index'].value == 'tzips') {
								return doc['network.keyword'].value + '|' + doc['address.keyword'].value
							} else if (doc['_index'].value == 'big_map_diffs') {
								return doc['key_hash.keyword'].value
							} else if (doc['_index'].value == 'tezos_domains') {
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
