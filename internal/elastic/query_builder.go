package elastic

import (
	"github.com/baking-bad/bcdhub/internal/helpers"
)

const (
	maxQuerySize = 10000
	minQuerySize = 0
)

var searchableInidices = []string{
	DocContracts,
	DocOperations,
	DocBigMapDiff,
	DocTZIP,
}

type qItem map[string]interface{}
type qList []interface{}

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

//nolint
func minimumShouldMatch(value int) qItem {
	return qItem{
		"minimum_should_match": value,
	}
}

func exists(field string) qItem {
	return qItem{
		"exists": qItem{
			"field": field,
		},
	}
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

func matchPhrase(key string, value interface{}) qItem {
	return qItem{
		"match_phrase": qItem{
			key: value,
		},
	}
}

func matchQ(key string, value interface{}) qItem {
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

func in(key string, value []string) qItem {
	return qItem{
		"terms": qItem{
			key: value,
		},
	}
}

type aggItem struct {
	name string
	body qItem
}

func aggs(items ...aggItem) qItem {
	body := qItem{}
	for i := range items {
		body[items[i].name] = items[i].body
	}
	return qItem{
		"aggs": body,
	}
}

func cardinality(field string) qItem {
	return qItem{
		"cardinality": qItem{
			"field": field,
		},
	}
}

func avg(field string) qItem {
	return qItem{
		"avg": qItem{
			"field": field,
		},
	}
}

func termsAgg(field string, size int64) qItem {
	t := qItem{
		"field": field,
	}
	if size > 0 {
		t["size"] = size
	}
	return qItem{
		"terms": t,
	}
}

func composite(size int64, items ...aggItem) qItem {
	body := make([]qItem, 0)
	for i := range items {
		body = append(body, qItem{
			items[i].name: items[i].body,
		})
	}
	return qItem{
		"composite": qItem{
			"sources": body,
			"size":    size,
		},
	}
}

//nolint
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

func min(field string) qItem {
	return qItem{
		"min": qItem{
			"field": field,
		},
	}
}

func sum(field string) qItem {
	return qItem{
		"sum": qItem{
			"field": field,
		},
	}
}

func count(field string) qItem {
	return qItem{
		"value_count": qItem{
			"field": field,
		},
	}
}

// nolint
func maxBucket(bucketsPath string) qItem {
	return qItem{
		"max_bucket": qItem{
			"buckets_path": bucketsPath,
		},
	}
}

// nolint
func minBucket(bucketsPath string) qItem {
	return qItem{
		"min_bucket": qItem{
			"buckets_path": bucketsPath,
		},
	}
}

func queryString(text string, fields []string) qItem {
	queryS := qItem{
		"query": text,
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

func (q qItem) Get(name string) qItem {
	if val, ok := q[name]; ok {
		if typ, ok := val.(qItem); ok {
			return typ
		}
		return nil
	}
	return nil
}

type base qItem

func newQuery() base {
	return base{}
}

func (q base) Size(size int64) base {
	if size != 0 {
		q["size"] = size
	}
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
	if from != 0 {
		q["from"] = from
	}
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

func (q base) SearchAfter(value []interface{}) base {
	q["search_after"] = value
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

func (q base) Get(name string) qItem {
	if val, ok := q[name]; ok {
		if typ, ok := val.(qItem); ok {
			return typ
		}
		return nil
	}
	return nil
}
