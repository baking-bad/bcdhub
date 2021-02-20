package noderpc

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"time"

	"github.com/baking-bad/bcdhub/internal/helpers"
	"github.com/baking-bad/bcdhub/internal/logger"
	jsoniter "github.com/json-iterator/go"
	"github.com/pkg/errors"
	"github.com/tidwall/gjson"
)

var json = jsoniter.ConfigCompatibleWithStandardLibrary

const (
	headBlock = "head"
)

func getBlockString(level int64) string {
	if level > 0 {
		return fmt.Sprintf("%d", level)
	}
	return headBlock
}

// NodeRPC -
type NodeRPC struct {
	baseURL string

	timeout    time.Duration
	retryCount int
}

// NewNodeRPC -
func NewNodeRPC(baseURL string, opts ...NodeOption) *NodeRPC {
	node := &NodeRPC{
		baseURL:    baseURL,
		timeout:    time.Second * 10,
		retryCount: 3,
	}

	for _, opt := range opts {
		opt(node)
	}
	return node
}

// NewWaitNodeRPC -
func NewWaitNodeRPC(baseURL string, opts ...NodeOption) *NodeRPC {
	node := NewNodeRPC(baseURL, opts...)

	for {
		if _, err := node.GetLevel(); err == nil {
			break
		}

		logger.Warning("Waiting node %s up 30 second...", baseURL)
		time.Sleep(time.Second * 30)
	}
	return node
}

func (rpc *NodeRPC) checkStatusCode(resp *http.Response, checkStatusCode bool) error {
	switch {
	case resp.StatusCode == http.StatusOK:
		return nil
	case resp.StatusCode > http.StatusInternalServerError:
		return NewNodeUnavailiableError(rpc.baseURL, resp.StatusCode)
	case checkStatusCode:
		return errors.Wrap(ErrInvalidStatusCode, fmt.Sprintf("%d", resp.StatusCode))
	default:
		return nil
	}
}

func (rpc *NodeRPC) parseResponse(resp *http.Response, checkStatusCode bool, response interface{}) error {
	if err := rpc.checkStatusCode(resp, checkStatusCode); err != nil {
		return err
	}
	return json.NewDecoder(resp.Body).Decode(response)
}

func (rpc *NodeRPC) getGJSONReponse(resp *http.Response, checkStatusCode bool) (result gjson.Result, err error) {
	if err := rpc.checkStatusCode(resp, checkStatusCode); err != nil {
		return result, err
	}

	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return result, err
	}

	if gjson.ValidBytes(b) {
		result = gjson.ParseBytes(b)
	} else {
		err = errors.Wrap(ErrInvalidNodeResponse, string(b))
	}
	return
}

func (rpc *NodeRPC) makeRequest(req *http.Request) (*http.Response, error) {
	client := http.Client{
		Timeout: rpc.timeout,
	}

	count := 0
	for ; count < rpc.retryCount; count++ {
		resp, err := client.Do(req)
		if err != nil {
			logger.Warning("Attempt #%d: %s", count+1, err.Error())
			continue
		}
		return resp, err
	}

	return nil, NewMaxRetryExceededError(rpc.baseURL)
}

func (rpc *NodeRPC) makeGetRequest(uri string) (*http.Response, error) {
	url := helpers.URLJoin(rpc.baseURL, uri)
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, errors.Errorf("makeGetRequest.NewRequest: %v", err)
	}
	return rpc.makeRequest(req)
}

func (rpc *NodeRPC) makePostRequest(uri string, data map[string]interface{}) (*http.Response, error) {
	url := helpers.URLJoin(rpc.baseURL, uri)

	bData, err := json.Marshal(data)
	if err != nil {
		return nil, errors.Errorf("makePostRequest.json.Marshal: %v", err)
	}

	req, err := http.NewRequest("POST", url, bytes.NewReader(bData))
	if err != nil {
		return nil, errors.Errorf("makePostRequest.NewRequest: %v", err)
	}
	return rpc.makeRequest(req)
}

//nolint
func (rpc *NodeRPC) get(uri string, response interface{}) error {
	resp, err := rpc.makeGetRequest(uri)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return rpc.parseResponse(resp, true, response)
}

//nolint
func (rpc *NodeRPC) getGJSON(uri string) (gjson.Result, error) {
	resp, err := rpc.makeGetRequest(uri)
	if err != nil {
		return gjson.Result{}, err
	}
	defer resp.Body.Close()

	return rpc.getGJSONReponse(resp, true)
}

//nolint
func (rpc *NodeRPC) post(uri string, data map[string]interface{}, checkStatusCode bool, response interface{}) error {
	resp, err := rpc.makePostRequest(uri, data)
	if err != nil {
		return NewMaxRetryExceededError(rpc.baseURL)
	}
	defer resp.Body.Close()

	return rpc.parseResponse(resp, checkStatusCode, response)
}

//nolint
func (rpc *NodeRPC) postGJSON(uri string, data map[string]interface{}, checkStatusCode bool) (gjson.Result, error) {
	resp, err := rpc.makePostRequest(uri, data)
	if err != nil {
		return gjson.Result{}, NewMaxRetryExceededError(rpc.baseURL)
	}
	defer resp.Body.Close()

	return rpc.getGJSONReponse(resp, checkStatusCode)
}

// GetHead - get head
func (rpc *NodeRPC) GetHead() (Header, error) {
	return rpc.GetHeader(0)
}

// GetLevel - get head level
func (rpc *NodeRPC) GetLevel() (int64, error) {
	var head struct {
		Level int64 `json:"level"`
	}
	if err := rpc.get("chains/main/blocks/head/header", &head); err != nil {
		return 0, err
	}
	return head.Level, nil
}

// GetHeader - get head for certain level
func (rpc *NodeRPC) GetHeader(level int64) (header Header, err error) {
	err = rpc.get(fmt.Sprintf("chains/main/blocks/%s/header", getBlockString(level)), &header)
	return
}

// GetLevelTime - get level time
func (rpc *NodeRPC) GetLevelTime(level int) (time.Time, error) {
	var head struct {
		Timestamp time.Time `json:"timestamp"`
	}
	if err := rpc.get(fmt.Sprintf("chains/main/blocks/%s/header", getBlockString(int64(level))), &head); err != nil {
		return time.Now(), err
	}
	return head.Timestamp.UTC(), nil
}

// GetScriptJSON -
func (rpc *NodeRPC) GetScriptJSON(address string, level int64) (gjson.Result, error) {
	return rpc.getGJSON(fmt.Sprintf("chains/main/blocks/%s/context/contracts/%s/script", getBlockString(level), address))
}

// GetScriptStorageJSON -
func (rpc *NodeRPC) GetScriptStorageJSON(address string, level int64) (gjson.Result, error) {
	return rpc.GetScriptJSON(address, level)
}

// GetContractBalance -
func (rpc *NodeRPC) GetContractBalance(address string, level int64) (int64, error) {
	var balanceStr string
	if err := rpc.get(fmt.Sprintf("chains/main/blocks/%s/context/contracts/%s/balance", getBlockString(level), address), &balanceStr); err != nil {
		return 0, err
	}
	return strconv.ParseInt(balanceStr, 10, 64)
}

// GetContractData -
func (rpc *NodeRPC) GetContractData(address string, level int64) (ContractData, error) {
	var response ContractData
	err := rpc.get(fmt.Sprintf("chains/main/blocks/%s/context/contracts/%s", getBlockString(level), address), &response)
	return response, err
}

// GetOperations -
func (rpc *NodeRPC) GetOperations(block int64) (gjson.Result, error) {
	return rpc.getGJSON(fmt.Sprintf("chains/main/blocks/%d/operations/3", block))
}

// GetContractsByBlock -
func (rpc *NodeRPC) GetContractsByBlock(block int64) ([]string, error) {
	if block != 1 {
		return nil, errors.Errorf("For less loading node RPC `block` value is only 1")
	}
	contracts := make([]string, 0)
	if err := rpc.get(fmt.Sprintf("chains/main/blocks/%d/context/contracts", block), &contracts); err != nil {
		return nil, err
	}
	return contracts, nil
}

// GetNetworkConstants -
func (rpc *NodeRPC) GetNetworkConstants(level int64) (constants Constants, err error) {
	err = rpc.get(fmt.Sprintf("chains/main/blocks/%s/context/constants", getBlockString(level)), &constants)
	return
}

// RunCode -
func (rpc *NodeRPC) RunCode(script, storage, input gjson.Result, chainID, source, payer, entrypoint, proto string, amount, gas int64) (gjson.Result, error) {
	data := map[string]interface{}{
		"script":   script.Value(),
		"storage":  storage.Value(),
		"input":    input.Value(),
		"amount":   fmt.Sprintf("%d", amount),
		"chain_id": chainID,
	}

	if chainID != "NetXm8tYqnMWky1" {
		data["balance"] = "0"
	}
	if gas != 0 {
		data["gas"] = fmt.Sprintf("%d", gas)
	}
	if source != "" {
		data["source"] = source
	}
	if payer != "" {
		data["payer"] = payer
	}
	if entrypoint != "" {
		data["entrypoint"] = entrypoint
	}

	return rpc.postGJSON("chains/main/blocks/head/helpers/scripts/run_code", data, false)
}

// RunOperation -
func (rpc *NodeRPC) RunOperation(chainID, branch, source, destination string, fee, gasLimit, storageLimit, counter, amount int64, parameters gjson.Result) (gjson.Result, error) {
	data := map[string]interface{}{
		"operation": map[string]interface{}{
			"branch":    branch,
			"signature": "sigUHx32f9wesZ1n2BWpixXz4AQaZggEtchaQNHYGRCoWNAXx45WGW2ua3apUUUAGMLPwAU41QoaFCzVSL61VaessLg4YbbP", // base58_encode(b'0' * 64, b'sig').decode()
			"contents": []interface{}{
				map[string]interface{}{
					"kind":          "transaction",
					"fee":           fmt.Sprintf("%d", fee),
					"counter":       fmt.Sprintf("%d", counter),
					"gas_limit":     fmt.Sprintf("%d", gasLimit),
					"storage_limit": fmt.Sprintf("%d", storageLimit),
					"source":        source,
					"destination":   destination,
					"amount":        fmt.Sprintf("%d", amount),
					"parameters":    parameters.Value(),
				},
			},
		},
		"chain_id": chainID,
	}

	return rpc.postGJSON("chains/main/blocks/head/helpers/scripts/run_operation", data, false)
}

// GetCounter -
func (rpc *NodeRPC) GetCounter(address string) (int64, error) {
	var counter string
	if err := rpc.get(fmt.Sprintf("chains/main/blocks/head/context/contracts/%s/counter", address), &counter); err != nil {
		return 0, err
	}
	return strconv.ParseInt(counter, 10, 64)
}

// GetCode -
func (rpc *NodeRPC) GetCode(address string, level int64) (gjson.Result, error) {
	contract, err := rpc.GetScriptJSON(address, level)
	if err != nil {
		return gjson.Result{}, err
	}

	return contract.Get("code"), nil
}
