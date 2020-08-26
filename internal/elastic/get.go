package elastic

import (
	"context"

	"github.com/elastic/go-elasticsearch/v8/esapi"
	"github.com/pkg/errors"
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

	result, err := e.getResponse(resp)
	if err != nil {
		return err
	}
	if !result.Get("found").Bool() {
		return errors.Errorf("%s: %s %s", RecordNotFound, ret.GetIndex(), ret.GetID())
	}
	ret.ParseElasticJSON(result)
	return nil
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
	ctx := newScrollContext(e, query, defaultScrollSize)
	return ctx.get(output)
}
