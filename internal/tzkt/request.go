package tzkt

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"
)

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

func (t *TzKT) request(method, endpoint string, params map[string]string, response interface{}) error {
	uri := fmt.Sprintf("%s%s", t.Host, endpoint)

	req, err := http.NewRequest(method, uri, nil)
	if err != nil {
		return fmt.Errorf("[http.NewRequest] %s", err)
	}
	q := req.URL.Query()
	for key, value := range params {
		q.Add(key, value)
	}
	req.URL.RawQuery = q.Encode()
	req.Header.Set("User-Agent", userAgent)

	// log.Println(req.URL)
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
		return fmt.Errorf("Max HTTP request retry exceeded")
	}
	defer resp.Body.Close()

	dec := json.NewDecoder(resp.Body)
	if err = dec.Decode(response); err != nil {
		return fmt.Errorf("[json.Decode] %s", err)
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

// GetAccounts - returns account by filters
func (t *TzKT) GetAccounts(kind string, page, limit int64) (resp []Account, err error) {
	params := map[string]string{}
	if kind != "" && kind != ContractKindAll {
		params["kind"] = kind
	}
	if limit == 0 {
		limit = 1000
	}
	params["n"] = fmt.Sprintf("%d", limit)
	params["p"] = fmt.Sprintf("%d", page)

	err = t.request("GET", "contracts", params, &resp)
	return
}

// GetContractByAddress - returns contract by address
func (t *TzKT) GetContractByAddress(address string) (resp Account, err error) {
	err = t.request("GET", fmt.Sprintf("contracts/%s", address), nil, &resp)
	return
}

// GetContractOperationBlocks -
func (t *TzKT) GetContractOperationBlocks(offset, limit int64) (resp []int64, err error) {
	params := map[string]string{}
	if limit == 0 {
		limit = 10000
	}

	params["limit"] = fmt.Sprintf("%d", limit)
	params["offset.cr"] = fmt.Sprintf("%d", offset)

	err = t.request("GET", "blocks/levels", params, &resp)
	return
}

// GetAliases - returns address aliases
func (t *TzKT) GetAliases() (resp []Alias, err error) {
	err = t.request("GET", "suggest/accounts", nil, &resp)
	return
}

// GetAllContractOperationBlocks -
func (t *TzKT) GetAllContractOperationBlocks() ([]int64, error) {
	offset := int64(0)
	resp := make([]int64, 0)
	end := false
	for !end {
		levels, err := t.GetContractOperationBlocks(offset, 0)
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
