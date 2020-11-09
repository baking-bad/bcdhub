package elastic

import (
	"context"

	"github.com/elastic/go-elasticsearch/v8/esapi"
)

// GetByID -
func (e *Elastic) GetByID(ret Model) error {
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
	if err := e.getResponse(resp, &response); err != nil {
		return err
	}
	if !response.Found {
		return NewRecordNotFoundError(ret.GetIndex(), ret.GetID(), nil)
	}
	return json.Unmarshal(response.Source, ret)
}

// GetByIDs -
func (e *Elastic) GetByIDs(output interface{}, ids ...string) error {
	query := newQuery().Query(
		qItem{
			"ids": qItem{
				"values": ids,
			},
		},
	)
	return e.getAllByQuery(query, output)
}

// GetAll -
func (e *Elastic) GetAll(output interface{}) error {
	return e.getAllByQuery(newQuery(), output)
}

// GetByNetwork -
func (e *Elastic) GetByNetwork(network string, output interface{}) error {
	query := newQuery().Query(
		boolQ(
			must(
				matchPhrase("network", network),
			),
		),
	).Sort("level", "asc")
	return e.getAllByQuery(query, output)
}

// GetByNetworkWithSort -
func (e *Elastic) GetByNetworkWithSort(network, sortField, sortOrder string, output interface{}) error {
	query := newQuery().Query(
		boolQ(
			must(
				matchPhrase("network", network),
			),
		),
	).Sort(sortField, sortOrder)

	return e.getAllByQuery(query, output)
}

func (e *Elastic) getAllByQuery(query base, output interface{}) error {
	ctx := newScrollContext(e, query, 0, defaultScrollSize)
	return ctx.get(output)
}

type getCountAggResponse struct {
	Agg struct {
		Body struct {
			Buckets []Bucket `json:"buckets"`
		} `json:"body"`
	} `json:"aggregations"`
}

func (e *Elastic) getCountAgg(index []string, query base) (map[string]int64, error) {
	var response getCountAggResponse
	if err := e.query(index, query, &response); err != nil {
		return nil, err
	}

	counts := make(map[string]int64)
	for _, item := range response.Agg.Body.Buckets {
		counts[item.Key] = item.DocCount
	}
	return counts, nil
}
