package elastic

import (
	"fmt"
	"strings"

	"github.com/aopoltorzhicky/bcdhub/internal/models"
	"github.com/tidwall/gjson"
)

const (
	defaultSize = 10
)

func getFields(fields []string) ([]string, map[string]interface{}, error) {
	if len(fields) == 0 {
		return allFields, mapHighlights, nil
	}

	f := make([]string, 0)
	h := make(map[string]interface{})
	for i := range fields {
		if nf, ok := mapFields[fields[i]]; ok {
			f = append(f, nf)
			s := strings.Split(nf, "^")
			h[s[0]] = map[string]interface{}{}
		} else {
			return nil, nil, fmt.Errorf("Unknown field: %s", fields[i])
		}
	}
	return f, h, nil
}

func getNetworksFilter(networks []string) ([]qItem, error) {
	res := make([]qItem, 0)
	for i := range networks {
		if _, ok := supportedNetworks[networks[i]]; !ok {
			return nil, fmt.Errorf("Unsupported network: %s", networks[i])
		}
		res = append(res, matchPhrase("network", networks[i]))
	}
	return res, nil
}

func setDateFilter(mustItems []qItem, dateFrom, dateTo uint) []qItem {
	if dateFrom == 0 && dateTo == 0 {
		return mustItems
	}
	ts := qItem{}
	if dateFrom > 0 {
		ts["gte"] = dateFrom * 1000
	}
	if dateTo > 0 {
		ts["lte"] = dateTo * 1000
	}

	mustItems = append(mustItems, rangeQ("timestamp", ts))
	return mustItems
}

// SearchByText -
func (e *Elastic) SearchByText(text string, offset int64, fields, networks []string, dateFrom, dateTo uint, grouping bool) (SearchResult, error) {
	query := newQuery()

	mustItems := make([]qItem, 0)
	if text != "" {
		internalFields, highlights, err := getFields(fields)
		if err != nil {
			return SearchResult{}, err
		}
		mustItems = append(mustItems, queryString(text, internalFields))

		query.Highlights(highlights)
	}
	mustItems = setDateFilter(mustItems, dateFrom, dateTo)

	b := boolQ()
	if len(mustItems) > 0 {
		b.Get("bool").Extend(must(mustItems...))
	}

	if len(networks) > 0 {
		networksFilter, err := getNetworksFilter(networks)
		if err != nil {
			return SearchResult{}, err
		}
		b.Get("bool").Extend(
			should(networksFilter...),
		).Append("minimum_should_match", 1)
	}

	if grouping {
		th := topHits(5, "_score", "desc")
		th.Get("top_hits").Append("highlight", qItem{
			"fields": mapHighlights,
		})

		query.Add(
			aggs(
				"projects",
				qItem{
					"terms": qItem{
						"script": "if (doc.containsKey('fingerprint.parameter')) {return doc['fingerprint.parameter'].value + '|' + doc['fingerprint.storage'].value + '|' + doc['fingerprint.code'].value} else {return doc['hash.keyword'].value}",
						"size":   defaultSize + offset,
						"order": qItem{
							"bucketsSort": "desc",
						},
					},
					"aggs": qItem{
						"last":        th,
						"bucketsSort": max("timestamp"),
					},
				},
			),
		).Zero()
	} else {
		query.From(offset).Size(defaultSize)
	}

	query.Query(b)

	resp, err := e.query([]string{DocContracts, DocOperations}, query)
	if err != nil {
		return SearchResult{}, err
	}

	if !grouping {
		return SearchResult{
			Items: parseSearchResponse(resp),
			Time:  resp.Get("took").Int(),
			Count: resp.Get("hits.total.value").Int(),
		}, nil
	}
	return SearchResult{
		Items: parseSearchGroupingResponse(resp, defaultSize, offset),
		Time:  resp.Get("took").Int(),
		Count: resp.Get("hits.total.value").Int(),
	}, nil
}

func parseSearchResponse(data gjson.Result) []SearchItem {
	items := make([]SearchItem, 0)
	arr := data.Get("hits.hits").Array()
	for i := range arr {
		index := arr[i].Get("_index").String()
		switch index {
		case DocContracts:
			var c models.Contract
			c.ParseElasticJSON(arr[i])
			item := SearchItem{
				Type:  DocContracts,
				Value: c.Address,
				Body:  c,
			}
			items = append(items, item)
		case DocOperations:
			var op models.Operation
			op.ParseElasticJSON(arr[i])
			item := SearchItem{
				Type:  DocOperations,
				Value: op.Hash,
				Body:  op,
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
			if count > 4 {
				count = 4
			}
			searchItem.Group = &Group{
				Count: arr[i].Get("doc_count").Int(),
				Top:   make([]Top, count),
			}
		}

		for j, item := range arr[i].Get("last.hits.hits").Array() {
			index := item.Get("_index").String()
			searchItem.Type = index

			switch index {
			case DocContracts:
				if j == 0 {
					var c models.Contract
					c.ParseElasticJSON(item)
					searchItem.Body = c
					searchItem.Value = c.Address
				} else {
					searchItem.Group.Top[j-1] = Top{
						Key:     item.Get("_source.address").String(),
						Network: item.Get("_source.network").String(),
					}
				}
			case DocOperations:
				for j, item := range arr[i].Get("last.hits.hits").Array() {
					if j == 0 {
						var op models.Operation
						op.ParseElasticJSON(item)
						searchItem.Body = op
						searchItem.Value = op.Hash
					} else {
						searchItem.Group.Top[j-1] = Top{
							Key:     item.Get("_source.hash").String(),
							Network: item.Get("_source.network").String(),
						}
					}
				}
			default:
			}
		}
		items = append(items, searchItem)
	}
	return items
}
