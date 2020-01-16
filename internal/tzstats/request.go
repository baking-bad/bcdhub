package tzstats

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"time"
)

// TzStats -
type TzStats struct {
	baseURL string

	timeout    time.Duration
	retryCount int
}

// NewTzStats -
func NewTzStats(baseURL string) *TzStats {
	return &TzStats{
		baseURL:    baseURL,
		timeout:    time.Second * 10,
		retryCount: 3,
	}
}

func (api *TzStats) parseStatusCode(resp *http.Response) error {
	switch resp.StatusCode {
	case http.StatusOK:
		return nil
	default:
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return err
		}
		var errResp ErrorResponse
		if err := json.Unmarshal(body, &errResp); err != nil {
			return err
		}
		var message string
		for _, e := range errResp.Errors {
			message += fmt.Sprintf("%s: %s", e.Message, e.Detail)
		}
		return fmt.Errorf("[ERROR] %s (%d) %s", resp.Status, resp.StatusCode, message)
	}
}

func (api *TzStats) get(uri string, params map[string]string, ret interface{}) error {
	url := fmt.Sprintf("%s/%s", api.baseURL, uri)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return err
	}
	q := req.URL.Query()
	for key, value := range params {
		q.Add(key, value)
	}
	req.URL.RawQuery = q.Encode()

	client := http.Client{Timeout: api.timeout}
	var resp *http.Response
	count := 0
	for ; count < api.retryCount; count++ {
		if resp, err = client.Do(req); err != nil {
			log.Printf("Attempt #%d: %s", count+1, err.Error())
			continue
		}
		break
	}

	if count == api.retryCount {
		return errors.New("Max HTTP request retry exceeded")
	}
	defer resp.Body.Close()

	if err = api.parseStatusCode(resp); err != nil {
		return err
	}

	return parseJSON(resp.Body, ret)
}

// SetTimeout - set request timeout
func (api *TzStats) SetTimeout(timeout time.Duration) {
	api.timeout = timeout
}

// GetTable -
func (api *TzStats) GetTable(table string, params map[string]string, response interface{}) error {
	return api.get(table, params, response)
}

// TableSnapshot - List network-wide staking status across all bakers and delegators at snapshot blocks. this table contains all snapshots regardless of them beeing later chosen as cycle snapshot or not.
func (api *TzStats) TableSnapshot(params map[string]string) (response TableResponse, err error) {
	err = api.GetTable("tables/snapshot", params, &response)
	return response, err
}

// TableAccount - List information about the most recent state of implicit and smart contract accounts.
func (api *TzStats) TableAccount(params map[string]string) (response TableResponse, err error) {
	err = api.GetTable("tables/account", params, &response)
	return response, err
}

// TableOperation - List detailed information about operations.
func (api *TzStats) TableOperation(params map[string]string) (response TableResponse, err error) {
	err = api.GetTable("tables/op", params, &response)
	return response, err
}

// TableIncome -
func (api *TzStats) TableIncome(params map[string]string) (response TableResponse, err error) {
	err = api.GetTable("tables/income", params, &response)
	return response, err
}

// Table - Select table for request
func (api *TzStats) Table(name string) *Query {
	return &Query{
		selectedTable: fmt.Sprintf("tables/%s", name),
		params:        make(map[string]string),
		api:           api,
	}
}

// Model - get table and column names from struct tag `tzstats`
func (api *TzStats) Model(s interface{}) *Query {
	params := make(map[string]string)
	if _, ok := params["columns"]; !ok {
		params["columns"] = getColumns(s)
	}
	return &Query{
		selectedTable: fmt.Sprintf("tables/%s", s.(Table).Name()),
		params:        params,
		api:           api,
	}
}
