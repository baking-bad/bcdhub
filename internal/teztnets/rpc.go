package teztnets

import (
	"context"
	"encoding/json"
	"net/http"
	"net/url"

	"github.com/pkg/errors"
)

// RPC -
type RPC struct {
	client *http.Client
	url    *url.URL
}

// New -
func New(baseURL string) (*RPC, error) {
	u, err := url.Parse(baseURL)
	if err != nil {
		return nil, err
	}
	t := http.DefaultTransport.(*http.Transport).Clone()
	t.MaxIdleConns = 100
	t.MaxConnsPerHost = 100
	t.MaxIdleConnsPerHost = 100

	client := &http.Client{
		Transport: t,
	}
	return &RPC{
		client: client,
		url:    u,
	}, nil
}

func (rpc *RPC) get(ctx context.Context, path string, output any) error {
	newUrl := rpc.url.JoinPath(path)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, newUrl.String(), nil)
	if err != nil {
		return err
	}

	response, err := rpc.client.Do(req)
	if err != nil {
		return err
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return errors.Errorf("invalid status code: %d", response.StatusCode)
	}

	err = json.NewDecoder(response.Body).Decode(output)
	return err
}

// Teztnets -
func (rpc *RPC) Teztnets(ctx context.Context) (info Info, err error) {
	info = make(Info)
	err = rpc.get(ctx, "teztnets.json", &info)
	return
}
