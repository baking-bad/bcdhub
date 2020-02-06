package elastic

import (
	"fmt"
	"strings"
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

func getNetworksFilter(networks []string) ([]map[string]interface{}, error) {
	res := make([]map[string]interface{}, 0)
	for i := range networks {
		if _, ok := supportedNetworks[networks[i]]; !ok {
			return nil, fmt.Errorf("Unsupported network: %s", networks[i])
		}
		res = append(res, map[string]interface{}{
			"match_phrase": map[string]interface{}{
				"network": networks[i],
			},
		})
	}
	return res, nil
}

func setDateFilter(must []map[string]interface{}, dateFrom, dateTo uint) []map[string]interface{} {
	if dateFrom <= 0 && dateTo <= 0 {
		return must
	}
	ts := map[string]interface{}{}
	if dateFrom > 0 {
		ts["gte"] = dateFrom * 1000
	}
	if dateTo > 0 {
		ts["lte"] = dateTo * 1000
	}

	must = append(must, map[string]interface{}{
		"range": map[string]interface{}{
			"timestamp": ts,
		},
	})
	return must
}

// SearchByText -
func (e *Elastic) SearchByText(text string, offset int64, fields, networks []string, dateFrom, dateTo uint) (SearchResult, error) {
	query := map[string]interface{}{
		"_source": map[string]interface{}{
			"excludes": []string{"hash"},
		},
		"size": 10,
		"from": offset,
	}

	networksFilter, err := getNetworksFilter(networks)
	if err != nil {
		return SearchResult{}, err
	}

	must := []map[string]interface{}{}
	if text != "" {
		internalFields, highlights, err := getFields(fields)
		if err != nil {
			return SearchResult{}, err
		}
		must = append(must, map[string]interface{}{
			"query_string": map[string]interface{}{
				"query":  fmt.Sprintf("*%s*", text),
				"fields": internalFields,
			},
		})
		query["highlight"] = map[string]interface{}{
			"fields": highlights,
		}
	}
	must = setDateFilter(must, dateFrom, dateTo)

	b := map[string]interface{}{
		"must": must,
	}
	if len(networks) > 0 {
		b["should"] = networksFilter
		b["minimum_should_match"] = 1
	}
	query["query"] = map[string]interface{}{
		"bool": b,
	}
	res, err := e.query(DocContracts, query)
	if err != nil {
		return SearchResult{}, err
	}
	return SearchResult{
		Contracts: parseContarcts(res),
		Time:      res.Get("took").Int(),
		Count:     res.Get("hits.total.value").Int(),
	}, nil
}
