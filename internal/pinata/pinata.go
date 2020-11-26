package pinata

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/baking-bad/bcdhub/internal/helpers"
)

const baseURL = "https://api.pinata.cloud/"

// Service -
type Service interface {
	PinList() (PinList, error)
	PinJSONToIPFS(data io.Reader) (PinJSONResponse, error)
}

// Pinata -
type Pinata struct {
	client       http.Client
	apiKey       string
	apiSecretKey string
}

// New -
func New(key, secretKey string, timeout time.Duration) *Pinata {
	return &Pinata{
		client: http.Client{
			Timeout: timeout,
		},
		apiKey:       key,
		apiSecretKey: secretKey,
	}
}

// PinList - https://pinata.cloud/documentation#PinList
func (p *Pinata) PinList() (PinList, error) {
	var ret PinList
	return ret, p.request("GET", "data/pinList", nil, make(map[string]string), &ret)
}

// PinJSONToIPFS - https://pinata.cloud/documentation#PinJSONToIPFS
func (p *Pinata) PinJSONToIPFS(body io.Reader) (PinJSONResponse, error) {
	var ret PinJSONResponse
	return ret, p.request("POST", "pinning/pinJSONToIPFS", body, make(map[string]string), &ret)
}

// UnPin - https://pinata.cloud/documentation#Unpin
func (p *Pinata) UnPin(hash string) error {
	var ret interface{}
	if err := p.request("DELETE", fmt.Sprintf("pinning/unpin/%s", hash), nil, make(map[string]string), &ret); err != nil {
		return err
	}

	fmt.Println(ret.(string))

	return nil
}

func (p *Pinata) request(method, endpoint string, body io.Reader, params map[string]string, response interface{}) error {
	url := helpers.URLJoin(baseURL, endpoint)

	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return fmt.Errorf("request http.NewRequest error %w", err)
	}

	q := req.URL.Query()
	for key, value := range params {
		q.Add(key, value)
	}
	req.URL.RawQuery = q.Encode()

	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("pinata_api_key", p.apiKey)
	req.Header.Add("pinata_secret_api_key", p.apiSecretKey)

	resp, err := p.client.Do(req)
	if err != nil {
		return fmt.Errorf("pinata request error %s %s %v %w", method, endpoint, params, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		return json.NewDecoder(resp.Body).Decode(response)
	}

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	return fmt.Errorf("%s", string(data))
}
