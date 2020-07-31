package elastic

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/baking-bad/bcdhub/internal/helpers"
	"github.com/baking-bad/bcdhub/internal/models"
	"github.com/tidwall/gjson"
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
			return nil, nil, fmt.Errorf("Unknown field: %s", field)
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
				return "", fmt.Errorf("Invalid type for 'from' filter (wait string): %T", v)
			}
			if val != "" {
				builder.WriteString(fmt.Sprintf("timestamp:{%s TO *}", val))
			}
		case "to":
			val, ok := v.(string)
			if !ok {
				return "", fmt.Errorf("Invalid type for 'to' filter (wait string): %T", v)
			}
			if val != "" {
				builder.WriteString(fmt.Sprintf("timestamp:{* TO %s}", val))
			}
		case "networks":
			val, ok := v.([]string)
			if !ok {
				return "", fmt.Errorf("Invalid type for 'network' filter (wait []string): %T", v)
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
				return "", fmt.Errorf("Invalid type for 'language' filter (wait []string): %T", v)
			}
			var str string
			if len(val) > 1 {
				str = fmt.Sprintf("language:(%s)", strings.Join(val, " OR "))
			} else {
				str = fmt.Sprintf("language:%s", val[0])
			}
			builder.WriteString(str)
		default:
			return "", fmt.Errorf("Unknown search filter: %s", k)
		}
	}
	return builder.String(), nil
}

func getSearchIndices(filters map[string]interface{}) ([]string, error) {
	if val, ok := filters["indices"]; ok {
		indices, ok := val.([]string)
		if !ok {
			return nil, fmt.Errorf("Invalid type for 'indices' filter (wait []string): %T", val)
		}
		for i := range indices {
			if !helpers.StringInArray(indices[i], searchableInidices) {
				return nil, fmt.Errorf("Invalid index name: %s", indices[i])
			}
		}
		delete(filters, "indices")
		return indices, nil
	}
	return searchableInidices, nil
}

// SearchByText -
func (e *Elastic) SearchByText(text string, offset int64, fields []string, filters map[string]interface{}, group bool) (SearchResult, error) {
	if text == "" {
		return SearchResult{}, fmt.Errorf("Empty search string. Please query something")
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

	resp, err := e.query(ctx.Indices, query)
	if err != nil {
		return SearchResult{}, err
	}

	var items []SearchItem
	if group {
		items = parseSearchGroupingResponse(resp, defaultSize, offset)
	} else {
		items = parseSearchResponse(resp)
	}

	return SearchResult{
		Items: items,
		Time:  resp.Get("took").Int(),
		Count: resp.Get("hits.total.value").Int(),
	}, nil
}

func parseHighlights(hit gjson.Result) map[string][]string {
	highlight := hit.Get("highlight").Map()
	res := make(map[string][]string, len(highlight))
	for k, v := range highlight {
		items := v.Array()
		res[k] = make([]string, len(items))
		for i, item := range items {
			res[k][i] = item.String()
		}
	}
	return res
}

func parseSearchResponse(data gjson.Result) []SearchItem {
	items := make([]SearchItem, 0)
	arr := data.Get("hits.hits").Array()
	for i := range arr {
		index := arr[i].Get("_index").String()
		highlights := parseHighlights(arr[i])
		switch index {
		case DocContracts:
			var c models.Contract
			c.ParseElasticJSON(arr[i])
			item := SearchItem{
				Type:       DocContracts,
				Value:      c.Address,
				Body:       c,
				Highlights: highlights,
			}
			items = append(items, item)
		case DocOperations:
			var op models.Operation
			op.ParseElasticJSON(arr[i])
			item := SearchItem{
				Type:       DocOperations,
				Value:      op.Hash,
				Body:       op,
				Highlights: highlights,
			}
			items = append(items, item)
		case DocBigMapDiff:
			var b models.BigMapDiff
			b.ParseElasticJSON(arr[i])
			item := SearchItem{
				Type:       DocBigMapDiff,
				Value:      b.KeyHash,
				Body:       b,
				Highlights: highlights,
			}
			items = append(items, item)
		default:
		}

	}
	return items
}

func parseSearchGroupingResponse(data gjson.Result, size, offset int64) []SearchItem {
	buckets := data.Get("aggregations.projects.buckets")
	if !buckets.Exists() {
		return nil
	}

	arr := buckets.Array()
	lArr := int64(len(arr))
	items := make([]SearchItem, 0)
	if offset > lArr {
		return items
	}
	arr = arr[offset:]
	for i := range arr {
		searchItem := SearchItem{}
		count := arr[i].Get("doc_count").Int()
		if count > 1 {
			searchItem.Group = &Group{
				Count: arr[i].Get("doc_count").Int(),
				Top:   make([]Top, 0),
			}
		}

		for j, item := range arr[i].Get("last.hits.hits").Array() {
			index := item.Get("_index").String()
			highlights := parseHighlights(item)
			searchItem.Type = index

			switch index {
			case DocContracts:
				if j == 0 {
					var c models.Contract
					c.ParseElasticJSON(item)
					searchItem.Body = c
					searchItem.Value = c.Address
					searchItem.Highlights = highlights
				} else {
					searchItem.Group.Top = append(searchItem.Group.Top, Top{
						Key:     item.Get("_source.address").String(),
						Network: item.Get("_source.network").String(),
					})
				}
			case DocOperations:
				for j, item := range arr[i].Get("last.hits.hits").Array() {
					if j == 0 {
						var op models.Operation
						op.ParseElasticJSON(item)
						searchItem.Body = op
						searchItem.Value = op.Hash
						searchItem.Highlights = highlights
					} else {
						searchItem.Group.Top = append(searchItem.Group.Top, Top{
							Key:     item.Get("_source.hash").String(),
							Network: item.Get("_source.network").String(),
						})
					}
				}
			case DocBigMapDiff:
				var b models.BigMapDiff
				b.ParseElasticJSON(arr[i].Get("last.hits.hits.0"))
				searchItem.Body = b
				searchItem.Value = b.KeyHash
				searchItem.Highlights = highlights
			default:
			}
		}
		items = append(items, searchItem)
	}
	return items
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
			"projects",
			qItem{
				"terms": qItem{
					"script": `
							if (doc.containsKey('fingerprint.parameter')) {
								return doc['fingerprint.parameter'].value + '|' + doc['fingerprint.storage'].value + '|' + doc['fingerprint.code'].value
							} else if (doc.containsKey('hash')) {
								return doc['hash.keyword'].value
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
		),
	).Zero()
	return query
}
