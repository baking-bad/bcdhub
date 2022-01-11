package elastic

import (
	"bytes"
	"context"
	stdJSON "encoding/json"
	"fmt"
	"io"

	"github.com/baking-bad/bcdhub/internal/search"
	"github.com/elastic/go-elasticsearch/v8/esapi"
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
	bulk := bytes.NewBuffer([]byte{})
	for i := range items {
		meta := fmt.Sprintf(`{"index":{"_id":"%s","_index":"%s"}}`, items[i].GetID(), items[i].GetIndex())
		if _, err := bulk.WriteString(meta); err != nil {
			return err
		}

		if err := bulk.WriteByte('\n'); err != nil {
			return err
		}

		data, err := json.Marshal(items[i])
		if err != nil {
			return err
		}

		if err := stdJSON.Compact(bulk, data); err != nil {
			return err
		}
		if err := bulk.WriteByte('\n'); err != nil {
			return err
		}

		if (i%1000 == 0 && i > 0) || i == len(items)-1 {
			if err := e.bulk(ctx, bulk); err != nil {
				return err
			}
			bulk.Reset()
		}
	}
	return nil
}
