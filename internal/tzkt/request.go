package tzkt

import (
	"fmt"
	"net/http"
	"time"

	"github.com/baking-bad/bcdhub/internal/helpers"
	"github.com/baking-bad/bcdhub/internal/logger"
	jsoniter "github.com/json-iterator/go"
	"github.com/pkg/errors"
)

var json = jsoniter.ConfigCompatibleWithStandardLibrary

// TzKT -
type TzKT struct {
	Host   string
	client http.Client

	retryCount int
}

// NewTzKT -
func NewTzKT(host string, timeout time.Duration) *TzKT {
	return &TzKT{
		Host: host,
		client: http.Client{
			Timeout: timeout,
		},

		retryCount: 3,
	}
}

//nolint
func (t *TzKT) request(method, endpoint string, params map[string]string, response interface{}) error {
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
			logger.Warning("Attempt #%d: %s", count+1, err.Error())
			continue
		}
		break
	}

	if count == t.retryCount {
		return errors.Errorf("Max HTTP request retry exceeded")
	}
	defer resp.Body.Close()

	return json.NewDecoder(resp.Body).Decode(response)
}

// GetHead - return head
func (t *TzKT) GetHead() (resp Head, err error) {
	err = t.request("GET", "head", nil, &resp)
	return
}

// GetAccounts - returns account by filters
func (t *TzKT) GetAccounts(kind string, page, limit int64) (resp []Account, err error) {
	params := map[string]string{}
	if kind != "" && kind != ContractKindAll {
		params["kind"] = kind
	}
	if limit == 0 {
		limit = 1000
	}
	params["limit"] = fmt.Sprintf("%d", limit)
	params["offset.pg"] = fmt.Sprintf("%d", page)

	err = t.request("GET", "contracts", params, &resp)
	return
}

// GetContractOperationBlocks -
func (t *TzKT) GetContractOperationBlocks(offset, limit int64, needSmartContracts, needDelegators bool) (resp []int64, err error) {
	params := map[string]string{}
	if limit == 0 {
		limit = 10000
	}

	params["limit"] = fmt.Sprintf("%d", limit)
	params["offset.cr"] = fmt.Sprintf("%d", offset)
	params["smartContracts"] = fmt.Sprintf("%v", needSmartContracts)
	params["delegatorContracts"] = fmt.Sprintf("%v", needDelegators)

	err = t.request("GET", "blocks/levels", params, &resp)
	return
}

// GetAliases - returns aliases map in format map[address]alias
func (t *TzKT) GetAliases() (map[string]string, error) {
	params := map[string]string{}

	params["limit"] = "10000"
	params["kind"] = "smart_contract"
	params["select.fields"] = "alias,address,creator,manager,delegate"

	var contracts []Contract
	if err := t.request("GET", "contracts", params, &contracts); err != nil {
		return nil, fmt.Errorf("request error %w", err)
	}

	aliases := make(map[string]string)
	for _, c := range contracts {
		if c.Alias != nil {
			aliases[c.Address] = *c.Alias
		}

		if c.Creator != nil {
			if c.Creator.Alias != nil && c.Creator.Address != nil {
				aliases[*c.Creator.Address] = *c.Creator.Alias
			}
		}

		if c.Manager != nil {
			if c.Manager.Alias != nil && c.Manager.Address != nil {
				aliases[*c.Manager.Address] = *c.Manager.Alias
			}
		}

		if c.Delegate != nil {
			if c.Delegate.Alias != nil && c.Delegate.Address != nil {
				aliases[*c.Delegate.Address] = *c.Delegate.Alias
			}
		}
	}
	return aliases, nil
}

// GetAllContractOperationBlocks -
func (t *TzKT) GetAllContractOperationBlocks() ([]int64, error) {
	offset := int64(0)
	resp := make([]int64, 0)
	end := false
	for !end {
		levels, err := t.GetContractOperationBlocks(offset, 0, true, true)
		if err != nil {
			return nil, err
		}
		for i := range levels {
			resp = append(resp, levels[i])
			if i == len(levels)-1 {
				offset = levels[i]
			}
		}
		end = len(levels) < 10000
	}
	return resp, nil
}

// GetProtocols -
func (t *TzKT) GetProtocols() (resp []Protocol, err error) {
	err = t.request("GET", "protocols", nil, &resp)
	return
}
