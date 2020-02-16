package elastic

import (
	"fmt"
	"strings"

	"github.com/aopoltorzhicky/bcdhub/internal/models"
	"github.com/tidwall/gjson"
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
	query := newQuery().From(offset)

	mustItems := make([]qItem, 0)
	if text != "" {
		internalFields, highlights, err := getFields(fields)
		if err != nil {
			return SearchResult{}, err
		}
		mustItems = append(mustItems, queryString(text, internalFields))

		query = query.Highlights(highlights)
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
		query = query.Add(
			aggs(
				"projects",
				qItem{
					"terms": qItem{
						"script": "doc['fingerprint.parameter'].value + '|' + doc['fingerprint.storage'].value + '|' + doc['fingerprint.code'].value",
						"size":   10000,
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
		query = query.Size(10)
	}

	query = query.Query(b)

	resp, err := e.query(DocContracts, query)
	if err != nil {
		return SearchResult{}, err
	}
	if !grouping {
		return SearchResult{
			Contracts: parseContracts(resp),
			Time:      resp.Get("took").Int(),
			Count:     resp.Get("hits.total.value").Int(),
		}, nil
	}
	return SearchResult{
		Contracts: parseGroupContracts(resp),
		Time:      resp.Get("took").Int(),
		Count:     resp.Get("hits.total.value").Int(),
	}, nil
}

func parseGroupContracts(data *gjson.Result) []models.Contract {
	buckets := data.Get("aggregations.projects.buckets")
	if !buckets.Exists() {
		return nil
	}
	contracts := make([]models.Contract, 0)
	arr := buckets.Array()
	for i := range arr {
		var c models.Contract
		for j, item := range arr[i].Get("last.hits.hits").Array() {
			if j == 0 {
				parseContractFromHit(item, &c)
			} else if j == 1 {
				c.Group = &models.Group{
					Count: arr[i].Get("doc_count").Int(),
					Top:   []string{item.Get("_source.address").String()},
				}
			} else {
				c.Group.Top = append(c.Group.Top, item.Get("_source.address").String())
			}
		}
		contracts = append(contracts, c)
	}
	return contracts
}
