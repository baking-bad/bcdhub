package elastic

import (
	"bytes"
	"context"
	"io"

	"github.com/baking-bad/bcdhub/internal/logger"
	"github.com/baking-bad/bcdhub/internal/search"
	"github.com/elastic/go-elasticsearch/v7/esapi"
	"github.com/elastic/go-elasticsearch/v7/esutil"
	"github.com/pkg/errors"
)

func (e *Elastic) bulk(ctx context.Context, body io.Reader) error {
	req := esapi.BulkRequest{
		Body:    body,
		Refresh: "true",
	}

	res, err := req.Do(ctx, e)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	var response BulkResponse
	if err := e.getResponse(res, &response); err != nil {
		return err
	}
	if response.Errors {
		return errors.Errorf("Bulk error: %s", string(response.Items))
	}
	return nil
}

// Save -
func (e *Elastic) Save(ctx context.Context, items []search.Data) error {
	if len(items) == 0 {
		return nil
	}

	for i := range items {
		data, err := json.Marshal(items[i])
		if err != nil {
			return err
		}
		if err := e.bulkIndexer.Add(ctx, esutil.BulkIndexerItem{
			Index:  items[i].GetIndex(),
			Action: "index",
			Body:   bytes.NewReader(data),
			OnFailure: func(ctx context.Context, item esutil.BulkIndexerItem, res esutil.BulkIndexerResponseItem, err error) {
				if err != nil {
					logger.Err(err)
				} else {
					logger.Error().Msgf("elastic bulk error: %s: %s", res.Error.Type, res.Error.Reason)
				}
			},
		}); err != nil {
			return err
		}
	}
	return nil
}
