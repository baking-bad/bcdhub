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
	UnPin(hash string) error
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
	response, err := p.request("GET", "data/pinList", nil, make(map[string]string))
	if err != nil {
		return ret, err
	}
	defer response.Body.Close()

	if response.StatusCode == http.StatusOK {
		return ret, json.NewDecoder(response.Body).Decode(&ret)
	}

	data, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return ret, err
	}

	return ret, fmt.Errorf("%s", string(data))
}

// PinJSONToIPFS - https://pinata.cloud/documentation#PinJSONToIPFS
func (p *Pinata) PinJSONToIPFS(data io.Reader) (PinJSONResponse, error) {
	var ret PinJSONResponse
	response, err := p.request("POST", "pinning/pinJSONToIPFS", data, make(map[string]string))
	if err != nil {
		return ret, err
	}
	defer response.Body.Close()

	if response.StatusCode == http.StatusOK {
		return ret, json.NewDecoder(response.Body).Decode(&ret)
	}

	b, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return ret, err
	}

	return ret, fmt.Errorf("%s", string(b))
}

// UnPin - https://pinata.cloud/documentation#Unpin
func (p *Pinata) UnPin(hash string) error {
	response, err := p.request("DELETE", fmt.Sprintf("pinning/unpin/%s", hash), nil, make(map[string]string))
	if err != nil {
		return err
	}
	defer response.Body.Close()

	if response.StatusCode == http.StatusOK {
		return nil
	}

	data, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return err
	}

	return fmt.Errorf("%s", string(data))
}

func (p *Pinata) request(method, endpoint string, body io.Reader, params map[string]string) (*http.Response, error) {
	url := helpers.URLJoin(baseURL, endpoint)

	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, fmt.Errorf("request http.NewRequest error %w", err)
	}

	q := req.URL.Query()
	for key, value := range params {
		q.Add(key, value)
	}
	req.URL.RawQuery = q.Encode()

	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("pinata_api_key", p.apiKey)
	req.Header.Add("pinata_secret_api_key", p.apiSecretKey)

	return p.client.Do(req)
}
