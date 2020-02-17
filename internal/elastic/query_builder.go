package elastic

import (
	"fmt"

	"github.com/aopoltorzhicky/bcdhub/internal/helpers"
)

const (
	maxQuerySize = 10000
	minQuerySize = 0
)

type qItem map[string]interface{}

func boolQ(items ...qItem) qItem {
	bq := qItem{}
	q := qItem{}
	for i := range items {
		for k, v := range items[i] {
			if helpers.StringInArray(k, []string{"must", "should", "filter", "must_not", "minimum_should_match"}) {
				q[k] = v
			}
		}
	}
	bq["bool"] = q
	return bq
}

func must(items ...qItem) qItem {
	return qItem{
		"must": items,
	}
}

func notMust(items ...qItem) qItem {
	return qItem{
		"must_not": items,
	}
}

func should(items ...qItem) qItem {
	return qItem{
		"should": items,
	}
}

func filter(items ...qItem) qItem {
	return qItem{
		"filter": items,
	}
}

func rangeQ(field string, orders ...qItem) qItem {
	q := qItem{}
	for i := range orders {
		for k, v := range orders[i] {
			if helpers.StringInArray(k, []string{"lt", "gt", "lte", "gte"}) {
				q[k] = v
			}
		}
	}
	return qItem{
		"range": qItem{
			field: q,
		},
	}
}

func matchAll() qItem {
	return qItem{
		"match_all": qItem{},
	}
}

func matchPhrase(key string, value interface{}) qItem {
	return qItem{
		"match_phrase": qItem{
			key: value,
		},
	}
}

func match(key string, value interface{}) qItem {
	return qItem{
		"match": qItem{
			key: value,
		},
	}
}

func term(key string, value interface{}) qItem {
	return qItem{
		"term": qItem{
			key: value,
		},
	}
}

func excludes(fields []string) qItem {
	return qItem{
		"excludes": fields,
	}
}

func includes(fields []string) qItem {
	return qItem{
		"includes": fields,
	}
}

func aggs(name string, item qItem) qItem {
	return qItem{
		"aggs": qItem{
			name: item,
		},
	}
}

func topHits(size int, sortField, order string) qItem {
	return qItem{
		"top_hits": qItem{
			"size": size,
			"sort": sort(sortField, order),
		},
	}
}

func sort(field, order string) qItem {
	return qItem{
		field: qItem{
			"order": order,
		},
	}
}

func max(field string) qItem {
	return qItem{
		"max": qItem{
			"field": field,
		},
	}
}

func queryString(text string, fields []string) qItem {
	queryS := qItem{
		"query": fmt.Sprintf("*%s*", text),
	}
	if len(fields) > 0 {
		queryS["fields"] = fields
	}
	return qItem{
		"query_string": queryS,
	}
}

func (q qItem) Append(key string, value interface{}) qItem {
	q[key] = value
	return q
}

func (q qItem) Extend(item qItem) qItem {
	for k, v := range item {
		q[k] = v
	}
	return q
}

func (q qItem) Get(key string) qItem {
	if v, ok := q[key]; ok {
		return v.(qItem)
	}
	return nil
}

type base qItem

func newQuery() base {
	return base{}
}

func (q base) Size(size int64) base {
	q["size"] = size
	return q
}

func (q base) All() base {
	q["size"] = maxQuerySize
	return q
}

func (q base) One() base {
	q["size"] = 1
	return q
}

func (q base) Zero() base {
	q["size"] = minQuerySize
	return q
}

func (q base) From(from int64) base {
	q["from"] = from
	return q
}

func (q base) Query(item qItem) base {
	q["query"] = item
	return q
}

func (q base) Sort(key, order string) base {
	q["sort"] = qItem{
		key: qItem{
			"order": order,
		},
	}
	return q
}

func (q base) Add(items ...qItem) base {
	for _, item := range items {
		for k, v := range item {
			q[k] = v
		}
	}
	return q
}

func (q base) Source(items ...qItem) base {
	qi := qItem{}
	for i := range items {
		for k, v := range items[i] {
			if helpers.StringInArray(k, []string{"excludes", "includes"}) {
				qi[k] = v
			}
		}
	}
	q["_source"] = qi
	return q
}

func (q base) Highlights(highlights qItem) base {
	q["highlight"] = qItem{
		"fields": highlights,
	}
	return q
}
