package tzkt

import (
	"fmt"
	"net/http"
	"time"

	"github.com/baking-bad/bcdhub/internal/helpers"
	"github.com/baking-bad/bcdhub/internal/logger"
	"github.com/pkg/errors"
)

// ServicesTzKT -
type ServicesTzKT struct {
	Network string
	Host    string
	client  http.Client

	retryCount int
}

// NewServicesTzKT -
func NewServicesTzKT(network, uri string, timeout time.Duration) *ServicesTzKT {
	return &ServicesTzKT{
		Host:    uri,
		Network: network,
		client: http.Client{
			Timeout: timeout,
		},
		retryCount: 3,
	}
}

//nolint
func (t *ServicesTzKT) request(method, endpoint string, params map[string]string, response interface{}) (err error) {
	uri := helpers.URLJoin(t.Host, endpoint)

	req, err := http.NewRequest(method, uri, nil)
	if err != nil {
		return errors.Errorf("[http.NewRequest] %s", err)
	}
	q := req.URL.Query()
	for key, value := range params {
		q.Add(key, value)
	}
	req.URL.RawQuery = q.Encode()
	req.Header.Set("User-Agent", userAgent)

	var resp *http.Response
	count := 0
	for ; count < t.retryCount; count++ {
		if resp, err = t.client.Do(req); err != nil {
			logger.Warning().Msgf("Attempt #%d: %s", count+1, err.Error())
			continue
		}
		break
	}

	if count == t.retryCount {
		return errors.Errorf("Max HTTP request retry exceeded")
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("got error response %v with code %d", resp, resp.StatusCode)
	}

	return json.NewDecoder(resp.Body).Decode(response)
}

// GetMempool -
func (t *ServicesTzKT) GetMempool(address string) ([]MempoolOperation, error) {
	operations := make([]MempoolOperation, 0)
	err := t.request("GET", fmt.Sprintf("mempool/%s", address), nil, &operations)
	return operations, err
}
