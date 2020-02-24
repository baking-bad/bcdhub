package tzkt

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"time"

	"github.com/pkg/errors"
	"github.com/tidwall/gjson"
)

// ServicesTzKT -
type ServicesTzKT struct {
	Network string
	Host    string
	client  http.Client

	retryCount int
}

// NewServicesTzKT -
func NewServicesTzKT(host, network string, timeout time.Duration) *ServicesTzKT {
	return &ServicesTzKT{
		Host:    host,
		Network: network,
		client: http.Client{
			Timeout: time.Duration(timeout) * time.Second,
		},

		retryCount: 3,
	}
}

func (t *ServicesTzKT) request(method, endpoint string, params map[string]string) (res gjson.Result, err error) {
	uri := fmt.Sprintf("%s/%s/v1/%s", t.Host, t.Network, endpoint)

	req, err := http.NewRequest(method, uri, nil)
	if err != nil {
		return res, errors.Wrap(err, "http.NewRequest")
	}
	q := req.URL.Query()
	for key, value := range params {
		q.Add(key, value)
	}
	req.URL.RawQuery = q.Encode()

	var resp *http.Response
	count := 0
	for ; count < t.retryCount; count++ {
		if resp, err = t.client.Do(req); err != nil {
			log.Printf("Attempt #%d: %s", count+1, err.Error())
			continue
		}
		break
	}

	if count == t.retryCount {
		return res, errors.New("Max HTTP request retry exceeded")
	}
	defer resp.Body.Close()

	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return
	}
	res = gjson.ParseBytes(b)
	return
}

// GetMempool -
func (t *ServicesTzKT) GetMempool(address string) (gjson.Result, error) {
	return t.request("GET", fmt.Sprintf("mempool/%s", address), nil)
}
