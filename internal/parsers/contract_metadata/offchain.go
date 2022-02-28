package contract_metadata

import (
	"context"
	"net/http"
	"time"

	"github.com/baking-bad/bcdhub/internal/helpers"
	"github.com/baking-bad/bcdhub/internal/models/contract_metadata"
	"github.com/baking-bad/bcdhub/internal/models/dapp"
	"github.com/baking-bad/bcdhub/internal/models/types"
	jsoniter "github.com/json-iterator/go"
)

var json = jsoniter.ConfigCompatibleWithStandardLibrary

// Offchain -
type Offchain struct {
	baseURL string
	client  *http.Client
}

// NewOffchain -
func NewOffchain(baseURL string) *Offchain {
	t := http.DefaultTransport.(*http.Transport).Clone()
	t.MaxIdleConns = 3
	t.MaxConnsPerHost = 3
	t.MaxIdleConnsPerHost = 3

	return &Offchain{
		client: &http.Client{
			Transport: t,
		},
		baseURL: baseURL,
	}
}

// GetDApps -
func (o *Offchain) GetDApps(ctx context.Context) (dapps []dapp.DApp, err error) {
	err = o.get(ctx, helpers.URLJoin(o.baseURL, "dapps_legacy.json"), &dapps)
	return
}

// GetContractMetadata -
func (o *Offchain) GetContractMetadata(ctx context.Context) (metadata []contract_metadata.ContractMetadata, err error) {
	if err = o.get(ctx, helpers.URLJoin(o.baseURL, "tzips_legacy.json"), &metadata); err != nil {
		return
	}

	for i := range metadata {
		metadata[i].OffChain = true
		metadata[i].Network = types.Mainnet
	}
	return
}

func (o *Offchain) get(ctx context.Context, url string, output interface{}) error {
	requestCtx, cancel := context.WithTimeout(ctx, time.Second*10)
	defer cancel()

	req, err := http.NewRequestWithContext(requestCtx, http.MethodGet, url, nil)
	if err != nil {
		return err
	}
	response, err := o.client.Do(req)
	if err != nil {
		return err
	}
	defer response.Body.Close()

	return json.NewDecoder(response.Body).Decode(output)
}
