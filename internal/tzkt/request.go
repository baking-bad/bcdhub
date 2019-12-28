package tzkt

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/pkg/errors"
)

// URLs
const (
	TzKTURLV1 = "https://api.tzkt.io/v1/"
)

// TzKT -
type TzKT struct {
	Host   string
	client http.Client
}

// NewTzKT -
func NewTzKT(host string, timeout time.Duration) *TzKT {
	return &TzKT{
		Host: host,
		client: http.Client{
			Timeout: time.Duration(timeout) * time.Second,
		},
	}
}

func (t *TzKT) request(method, endpoint string, params map[string]string, response interface{}) error {
	uri := fmt.Sprintf("%s%s", t.Host, endpoint)

	req, err := http.NewRequest(method, uri, nil)
	if err != nil {
		return err
	}
	q := req.URL.Query()
	for key, value := range params {
		q.Add(key, value)
	}
	req.URL.RawQuery = q.Encode()
	resp, err := t.client.Do(req)
	if err != nil {
		return errors.Wrapf(err, "[%s]: %s", method, req.URL.String())
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return errors.Wrap(err, "ioutil.ReadAll")
	}
	err = json.Unmarshal(body, &response)
	if err != nil {
		return errors.Wrap(err, "json.Unmarshal")
	}
	return nil
}

// GetHead - return head
func (t *TzKT) GetHead() (resp Head, err error) {
	err = t.request("GET", "head", nil, &resp)
	return
}

// GetOriginations - return originations
func (t *TzKT) GetOriginations(page, limit int64) (resp []Origination, err error) {
	if limit == 0 {
		limit = 1000
	}
	params := map[string]string{
		"p": fmt.Sprintf("%d", page),
		"n": fmt.Sprintf("%d", limit),
	}
	err = t.request("GET", "operations/originations", params, &resp)
	return
}

// GetOriginationsCount - return originations count
func (t *TzKT) GetOriginationsCount() (resp int64, err error) {
	err = t.request("GET", "operations/originations/count", nil, &resp)
	return
}

// GetSystemOperations - return system operations
func (t *TzKT) GetSystemOperations(page, limit int64) (resp []SystemOperation, err error) {
	if limit == 0 {
		limit = 1000
	}
	params := map[string]string{
		"p": fmt.Sprintf("%d", page),
		"n": fmt.Sprintf("%d", limit),
	}
	err = t.request("GET", "operations/system", params, &resp)
	return
}

// GetSystemOperationsCount - return system operations count
func (t *TzKT) GetSystemOperationsCount() (resp int64, err error) {
	err = t.request("GET", "operations/system/count", nil, &resp)
	return
}
