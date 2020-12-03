package elastic

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/baking-bad/bcdhub/internal/elastic/search"
	"github.com/pkg/errors"
)

const (
	defaultSize = 10
)

var ptrRegEx = regexp.MustCompile(`^ptr:\d+$`)
var sanitizeRegEx = regexp.MustCompile(`[\:]`)

type searchContext struct {
	Text       string
	Indices    []string
	Fields     []string
	Highlights qItem
	Offset     int64
}

func newSearchContext() searchContext {
	return searchContext{
		Fields:     make([]string, 0),
		Indices:    make([]string, 0),
		Highlights: make(qItem),
	}
}

func getFields(searchString string, filters map[string]interface{}, fields []string) ([]string, []string, qItem, error) {
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
	h := make(qItem)
	for _, score := range scores.Scores {
		s := strings.Split(score, "^")
		h[s[0]] = qItem{}
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

// searchByTextResponse -
type searchByTextResponse struct {
	Took int64     `json:"took"`
	Hits HitsArray `json:"hits"`
	Agg  struct {
		Projects struct {
			Buckets []struct {
				Bucket
				Last struct {
					Hits HitsArray `json:"hits"`
				} `json:"last"`
			} `json:"buckets"`
		} `json:"projects"`
	} `json:"aggregations"`
}

// SearchByText -
func (e *Elastic) SearchByText(text string, offset int64, fields []string, filters map[string]interface{}, group bool) (search.Result, error) {
	if text == "" {
		return search.Result{}, errors.Errorf("Empty search string. Please query something")
	}

	ctx, err := prepare(text, filters, fields)
	if err != nil {
		return search.Result{}, err
	}
	ctx.Offset = offset

	query := newQuery().Query(
		queryString(ctx.Text, ctx.Fields),
	)

	if group {
		query = grouping(ctx, query)
	}

	var response searchByTextResponse
	if err := e.query(ctx.Indices, query, &response); err != nil {
		return search.Result{}, err
	}

	var items []search.Item
	if group {
		items, err = parseSearchGroupingResponse(response, offset)
	} else {
		items, err = parseSearchResponse(response)
	}
	if err != nil {
		return search.Result{}, nil
	}

	return search.Result{
		Items: items,
		Time:  response.Took,
		Count: response.Hits.Total.Value,
	}, nil
}

func parseSearchResponse(response searchByTextResponse) ([]search.Item, error) {
	items := make([]search.Item, 0)
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
		case search.Item:
			items = append(items, t)
		case []search.Item:
			items = append(items, t...)
		}
	}
	return items, nil
}

func parseSearchGroupingResponse(response searchByTextResponse, offset int64) ([]search.Item, error) {
	if len(response.Agg.Projects.Buckets) == 0 {
		return make([]search.Item, 0), nil
	}

	arr := response.Agg.Projects.Buckets
	lArr := int64(len(arr))
	items := make([]search.Item, 0)
	if offset > lArr {
		return items, nil
	}
	arr = arr[offset:]
	for i := range arr {
		searchItem := search.Item{}
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
			switch t := val.(type) {
			case search.Item:
				if j == 0 {
					searchItem.Type = t.Type
					searchItem.Body = t.Body
					searchItem.Value = t.Value
					searchItem.Highlights = item.Highlight
				} else {
					searchItem.Group.Top = append(searchItem.Group.Top, search.Top{
						Key:     t.Value,
						Network: t.Network,
					})
				}
			case []search.Item:
				if j == 0 {
					if len(t) > 0 {
						searchItem.Type = t[0].Type
						searchItem.Body = t[0].Body
						searchItem.Value = t[0].Value
						searchItem.Highlights = item.Highlight
					}
					if len(t) > 1 {
						for k := range t[1:] {
							searchItem.Group.Top = append(searchItem.Group.Top, search.Top{
								Key:     t[k].Value,
								Network: t[k].Network,
							})
						}
					}
				} else {
					for k := range t {
						searchItem.Group.Top = append(searchItem.Group.Top, search.Top{
							Key:     t[k].Value,
							Network: t[k].Network,
						})
					}
				}
			}
		}
		items = append(items, searchItem)
	}
	return items, nil
}

func prepare(search string, filters map[string]interface{}, fields []string) (searchContext, error) {
	ctx := newSearchContext()

	if ptrRegEx.MatchString(search) {
		ctx.Text = strings.TrimPrefix(search, "ptr:")
		ctx.Indices = []string{DocBigMapDiff}
		ctx.Fields = []string{"ptr"}
	} else {
		internalFields, usingIndices, highlights, err := getFields(ctx.Text, filters, fields)
		if err != nil {
			return ctx, err
		}
		ctx.Indices = usingIndices
		ctx.Highlights = highlights
		ctx.Fields = internalFields
		ctx.Text = fmt.Sprintf("%s*", search)
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

func grouping(ctx searchContext, query base) base {
	topHits := qItem{
		"top_hits": qItem{
			"size": 1,
			"sort": qList{
				sort("_score", "desc"),
				qItem{"last_action": qItem{"order": "desc", "unmapped_type": "long"}},
				sort("timestamp", "desc"),
			},
			"highlight": qItem{
				"fields": ctx.Highlights,
			},
		},
	}

	query.Add(
		aggs(
			aggItem{
				"projects",
				qItem{
					"terms": qItem{
						"script": `
							if (doc['_index'].value == "contract") {
								return doc['fingerprint.parameter'].value + '|' + doc['fingerprint.storage'].value + '|' + doc['fingerprint.code'].value
							} else if (doc['_index'].value == 'operation') {
								return doc['hash.keyword'].value
							} else if (doc['_index'].value == 'tzip') {
								return doc['address.keyword'].value + '|' + doc['network.keyword'].value
							} else if (doc['_index'].value == 'bigmapdiff') {
								return doc['key_hash.keyword'].value
							} else if (doc['_index'].value == 'tezos_domain') {
								return doc['name.keyword'].value + '|' + doc['network.keyword'].value
							}`,
						"size": defaultSize + ctx.Offset,
						"order": qList{
							qItem{"bucket_score": "desc"},
							qItem{"bucket_time": "desc"},
						},
					},
					"aggs": qItem{
						"last": topHits,
						"bucket_score": qItem{
							"max": qItem{
								"script": "_score",
							},
						},
						"bucket_time": qItem{
							"max": qItem{
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
