package elastic

import (
	"context"

	"github.com/elastic/go-elasticsearch/v8/esapi"
	"github.com/tidwall/gjson"
)

// GetByID -
func (e *Elastic) GetByID(index, id string) (result gjson.Result, err error) {
	req := esapi.GetRequest{
		Index:      index,
		DocumentID: id,
	}
	resp, err := req.Do(context.Background(), e)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	result, err = e.getResponse(resp)
	return
}

// GetByIDs -
func (e *Elastic) GetByIDs(index string, ids []string) (result gjson.Result, err error) {
	query := newQuery().Query(
		qItem{
			"ids": qItem{
				"values": ids,
			},
		},
	)
	return e.query([]string{index}, query)
}
