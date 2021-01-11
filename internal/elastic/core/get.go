package core

import (
	"context"
	"strings"

	"github.com/baking-bad/bcdhub/internal/models"
	"github.com/elastic/go-elasticsearch/v8/esapi"
)

// GetByID -
func (e *Elastic) GetByID(ret models.Model) error {
	req := esapi.GetRequest{
		Index:      ret.GetIndex(),
		DocumentID: ret.GetID(),
	}
	resp, err := req.Do(context.Background(), e)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	var response GetResponse
	if err := e.GetResponse(resp, &response); err != nil {
		return err
	}
	if !response.Found {
		return NewRecordNotFoundError(ret.GetIndex(), ret.GetID())
	}
	return json.Unmarshal(response.Source, ret)
}

// GetByIDs -
func (e *Elastic) GetByIDs(output interface{}, ids ...string) error {
	query := NewQuery().Query(
		Item{
			"ids": Item{
				"values": ids,
			},
		},
	)
	return e.GetAllByQuery(query, output)
}

// GetAll -
func (e *Elastic) GetAll(output interface{}) error {
	return e.GetAllByQuery(NewQuery(), output)
}

// GetByNetwork -
func (e *Elastic) GetByNetwork(network string, output interface{}) error {
	query := NewQuery().Query(
		Bool(
			Must(
				MatchPhrase("network", network),
			),
		),
	).Sort("level", "asc")
	return e.GetAllByQuery(query, output)
}

// GetByNetworkWithSort -
func (e *Elastic) GetByNetworkWithSort(network, sortField, sortOrder string, output interface{}) error {
	query := NewQuery().Query(
		Bool(
			Must(
				MatchPhrase("network", network),
			),
		),
	).Sort(sortField, sortOrder)

	return e.GetAllByQuery(query, output)
}

// GetAllByQuery -
func (e *Elastic) GetAllByQuery(query Base, output interface{}) error {
	ctx := NewScrollContext(e, query, 0, defaultScrollSize)
	return ctx.Get(output)
}

type getCountAggResponse struct {
	Agg struct {
		Body struct {
			Buckets []Bucket `json:"buckets"`
		} `json:"body"`
	} `json:"aggregations"`
}

// GetCountAgg -
func (e *Elastic) GetCountAgg(index []string, query Base) (map[string]int64, error) {
	var response getCountAggResponse
	if err := e.Query(index, query, &response); err != nil {
		return nil, err
	}

	counts := make(map[string]int64)
	for _, item := range response.Agg.Body.Buckets {
		counts[item.Key] = item.DocCount
	}
	return counts, nil
}

// FiltersToQuery -
func FiltersToQuery(by map[string]interface{}) Base {
	matches := make([]Item, 0)

	for k, v := range by {
		if strings.HasSuffix(k, ".or") {
			field := strings.TrimSuffix(k, ".or")
			if field == "" {
				continue
			}
			shouldItems := make([]Item, 0)

			if val, ok := v.([]interface{}); ok {
				for _, v := range val {
					shouldItems = append(shouldItems, MatchPhrase(field, v))
				}
			} else {
				continue
			}

			matches = append(matches, Bool(Should(shouldItems...), MinimumShouldMatch(1)))
		} else {
			matches = append(matches, MatchPhrase(k, v))
		}

	}
	return NewQuery().Query(
		Bool(
			Must(matches...),
		),
	)
}
