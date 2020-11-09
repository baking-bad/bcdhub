package elastic

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/baking-bad/bcdhub/internal/helpers"
	"github.com/baking-bad/bcdhub/internal/models"
	"github.com/pkg/errors"
)

const (
	defaultSize = 10
)

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

func getMapFields(allFields []string) map[string]string {
	res := make(map[string]string)
	for _, f := range allFields {
		str := strings.Split(f, "^")
		res[str[0]] = f
	}
	return res
}

func getHighlights(allFields []string) qItem {
	res := make(qItem)
	for _, f := range allFields {
		str := strings.Split(f, "^")
		res[str[0]] = qItem{}
	}
	return res
}

func getFields(search string, fields []string) ([]string, qItem, error) {
	if len(fields) == 0 {
		allFields, err := GetSearchScores(search, searchableInidices)
		if err != nil {
			return nil, nil, err
		}
		highlights := getHighlights(allFields)
		return allFields, highlights, nil
	}

	allFields, err := GetSearchScores(search, fields)
	if err != nil {
		return nil, nil, err
	}
	mapFields := getMapFields(allFields)

	f := make([]string, 0)
	h := make(qItem)
	for _, field := range fields {
		if nf, ok := mapFields[field]; ok {
			f = append(f, nf)
			s := strings.Split(nf, "^")
			h[s[0]] = qItem{}
		} else {
			return nil, nil, errors.Errorf("Unknown field: %s", field)
		}
	}
	return f, h, nil
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

func getSearchIndices(filters map[string]interface{}) ([]string, error) {
	if val, ok := filters["indices"]; ok {
		indices, ok := val.([]string)
		if !ok {
			return nil, errors.Errorf("Invalid type for 'indices' filter (wait []string): %T", val)
		}
		for i := range indices {
			if !helpers.StringInArray(indices[i], searchableInidices) {
				return nil, errors.Errorf("Invalid index name: %s", indices[i])
			}
		}
		delete(filters, "indices")
		return indices, nil
	}
	return searchableInidices, nil
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
func (e *Elastic) SearchByText(text string, offset int64, fields []string, filters map[string]interface{}, group bool) (SearchResult, error) {
	if text == "" {
		return SearchResult{}, errors.Errorf("Empty search string. Please query something")
	}

	ctx, err := prepare(text, filters, fields)
	if err != nil {
		return SearchResult{}, err
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
		return SearchResult{}, err
	}

	var items []SearchItem
	if group {
		items, err = parseSearchGroupingResponse(response, offset)
	} else {
		items, err = parseSearchResponse(response)
	}
	if err != nil {
		return SearchResult{}, nil
	}

	return SearchResult{
		Items: items,
		Time:  response.Took,
		Count: response.Hits.Total.Value,
	}, nil
}

func parseSearchResponse(response searchByTextResponse) ([]SearchItem, error) {
	items := make([]SearchItem, 0)
	arr := response.Hits.Hits
	for i := range arr {
		switch arr[i].Index {
		case DocContracts:
			var c models.Contract
			if err := json.Unmarshal(arr[i].Source, &c); err != nil {
				return nil, err
			}
			item := SearchItem{
				Type:       DocContracts,
				Value:      c.Address,
				Body:       c,
				Highlights: arr[i].Highlights,
			}
			items = append(items, item)
		case DocOperations:
			var op models.Operation
			if err := json.Unmarshal(arr[i].Source, &op); err != nil {
				return nil, err
			}
			item := SearchItem{
				Type:       DocOperations,
				Value:      op.Hash,
				Body:       op,
				Highlights: arr[i].Highlights,
			}
			items = append(items, item)
		case DocBigMapDiff:
			var b models.BigMapDiff
			if err := json.Unmarshal(arr[i].Source, &b); err != nil {
				return nil, err
			}
			item := SearchItem{
				Type:       DocBigMapDiff,
				Value:      b.KeyHash,
				Body:       b,
				Highlights: arr[i].Highlights,
			}
			items = append(items, item)
		case DocTZIP:
			var token models.TZIP
			if err := json.Unmarshal(arr[i].Source, &token); err != nil {
				return nil, err
			}
			item := SearchItem{
				Type:       DocBigMapDiff,
				Value:      token.Address,
				Body:       token,
				Highlights: arr[i].Highlights,
			}
			items = append(items, item)
		default:
		}

	}
	return items, nil
}

func parseSearchGroupingResponse(response searchByTextResponse, offset int64) ([]SearchItem, error) {
	if len(response.Agg.Projects.Buckets) == 0 {
		return nil, nil
	}

	arr := response.Agg.Projects.Buckets
	lArr := int64(len(arr))
	items := make([]SearchItem, 0)
	if offset > lArr {
		return items, nil
	}
	arr = arr[offset:]
	for i := range arr {
		searchItem := SearchItem{}
		if arr[i].DocCount > 1 {
			searchItem.Group = &Group{
				Count: arr[i].DocCount,
				Top:   make([]Top, 0),
			}
		}

		for j, item := range arr[i].Last.Hits.Hits {
			searchItem.Type = item.Index

			switch item.Index {
			case DocContracts:
				var c models.Contract
				if err := json.Unmarshal(item.Source, &c); err != nil {
					return nil, err
				}
				if j == 0 {
					searchItem.Body = c
					searchItem.Value = c.Address
					searchItem.Highlights = item.Highlights
				} else {
					searchItem.Group.Top = append(searchItem.Group.Top, Top{
						Key:     c.Address,
						Network: c.Network,
					})
				}
			case DocOperations:
				var op models.Operation
				if err := json.Unmarshal(item.Source, &op); err != nil {
					return nil, err
				}
				if j == 0 {
					searchItem.Body = op
					searchItem.Value = op.Hash
					searchItem.Highlights = item.Highlights
				} else {
					searchItem.Group.Top = append(searchItem.Group.Top, Top{
						Key:     op.Hash,
						Network: op.Network,
					})
				}
			case DocBigMapDiff:
				var b models.BigMapDiff
				if err := json.Unmarshal(item.Source, &b); err != nil {
					return nil, err
				}
				searchItem.Body = b
				searchItem.Value = b.KeyHash
				searchItem.Highlights = item.Highlights
			case DocTZIP:
				var token models.TZIP
				if err := json.Unmarshal(item.Source, &token); err != nil {
					return nil, err
				}
				searchItem.Body = token
				searchItem.Value = token.Address
				searchItem.Highlights = item.Highlights
			default:
			}
		}
		items = append(items, searchItem)
	}
	return items, nil
}

func prepare(search string, filters map[string]interface{}, fields []string) (searchContext, error) {
	ctx := newSearchContext()

	needEscape := true
	re := regexp.MustCompile(`^ptr:\d+$`)
	if re.MatchString(search) {
		ctx.Text = strings.TrimPrefix(search, "ptr:")
		ctx.Indices = []string{DocBigMapDiff}
		ctx.Fields = []string{"ptr"}
		needEscape = false
	} else {
		sanitized := `[\+\-\=\&\|\>\<\!\(\)\{\}\[\]\^\"\~\*\?\:\\\/]`
		re = regexp.MustCompile(sanitized)
		ctx.Text = re.ReplaceAllString(search, "\\${1}")

		indices, err := getSearchIndices(filters)
		if err != nil {
			return ctx, err
		}
		internalFields, highlights, err := getFields(ctx.Text, fields)
		if err != nil {
			return ctx, err
		}
		ctx.Indices = indices
		ctx.Highlights = highlights
		ctx.Fields = internalFields
	}

	if needEscape {
		ctx.Text = fmt.Sprintf("*%s*", ctx.Text)
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
							if (doc.containsKey('fingerprint.parameter')) {
								return doc['fingerprint.parameter'].value + '|' + doc['fingerprint.storage'].value + '|' + doc['fingerprint.code'].value
							} else if (doc.containsKey('hash')) {
								return doc['hash.keyword'].value
							} else if (doc.containsKey('token_id')) {
								return doc['contract.keyword'].value + '|' + doc['network.keyword'].value + '|' + doc['token_id'].value
							} else return doc['key_hash.keyword'].value`,
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
