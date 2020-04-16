package tzkt

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"time"

	"github.com/tidwall/gjson"
)

const (
	tzKTServicesURL         = "https://services.tzkt.io"
	tzKTServicesBabylonURL  = "https://services.tzkt.io/babylonnet"
	tzKTServicesCarthageURL = "https://services.tzkt.io/carthagenet"
	tzKTServicesZeroURL     = "https://services.tzkt.io/zeronet"
)

// ServicesTzKT -
type ServicesTzKT struct {
	Network string
	Host    string
	client  http.Client

	retryCount int
}

// NewServicesTzKT -
func NewServicesTzKT(network string, timeout time.Duration) *ServicesTzKT {
	var host string
	switch network {
	case "babylonnet":
		host = tzKTServicesBabylonURL
	case "carthagenet":
		host = tzKTServicesCarthageURL
	case "zeronet":
		host = tzKTServicesZeroURL
	case "mainnet":
		host = tzKTServicesURL
	default:
		panic(fmt.Sprintf("TzKT does not support %s", network))
	}
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
	uri := fmt.Sprintf("%s/v1/%s", t.Host, endpoint)

	req, err := http.NewRequest(method, uri, nil)
	if err != nil {
		return res, fmt.Errorf("[http.NewRequest] %s", err)
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
			log.Printf("Attempt #%d: %s", count+1, err.Error())
			continue
		}
		break
	}

	if count == t.retryCount {
		return res, fmt.Errorf("Max HTTP request retry exceeded")
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
