package contract_metadata

import (
	"context"
	"net/http"
	"time"

	"github.com/baking-bad/bcdhub/internal/bcd/consts"
	"github.com/baking-bad/bcdhub/internal/helpers"
	"github.com/baking-bad/bcdhub/internal/models/account"
	"github.com/baking-bad/bcdhub/internal/models/contract_metadata"
	"github.com/baking-bad/bcdhub/internal/models/dapp"
	"github.com/baking-bad/bcdhub/internal/models/tokenmetadata"
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

// Metadata
type contractMetadataLegacy struct {
	contract_metadata.ContractMetadata
	Tokens struct {
		Static []struct {
			Name     string                 `json:"name"`
			Symbol   string                 `json:"symbol,omitempty"`
			Decimals *int64                 `json:"decimals,omitempty"`
			TokenID  uint64                 `json:"token_id"`
			Extras   map[string]interface{} `json:"extras"`
		} `json:"static,omitempty"`
	} `json:"tokens"`
}

// ContractMetadata -
type ContractMetadata struct {
	Accounts  []account.Account
	Contracts []contract_metadata.ContractMetadata
	Tokens    []tokenmetadata.TokenMetadata
}

// GetContractMetadata -
func (o *Offchain) GetContractMetadata(ctx context.Context) (ContractMetadata, error) {
	var contractMetadata []contractMetadataLegacy
	if err := o.get(ctx, helpers.URLJoin(o.baseURL, "tzips_legacy.json"), &contractMetadata); err != nil {
		return ContractMetadata{}, err
	}

	result := ContractMetadata{
		Tokens:    make([]tokenmetadata.TokenMetadata, 0),
		Contracts: make([]contract_metadata.ContractMetadata, len(contractMetadata)),
		Accounts:  make([]account.Account, 0),
	}
	for i := range contractMetadata {
		result.Contracts[i] = contractMetadata[i].ContractMetadata
		result.Contracts[i].OffChain = true

		result.Accounts = append(result.Accounts, account.Account{
			Address: contractMetadata[i].Address,
			Alias:   contractMetadata[i].Name,
			Type:    types.NewAccountType(contractMetadata[i].Address),
		})

		for _, token := range contractMetadata[i].Tokens.Static {
			result.Tokens = append(result.Tokens, tokenmetadata.TokenMetadata{
				Contract:  result.Contracts[i].Address,
				TokenID:   token.TokenID,
				Decimals:  token.Decimals,
				Symbol:    token.Symbol,
				Extras:    token.Extras,
				Timestamp: consts.BeginningOfTime,
				Level:     0,
				Name:      token.Name,
			})
		}
	}
	return result, nil
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
