package noderpc

import (
	"bytes"
	stdJSON "encoding/json"
	"fmt"
	"io/ioutil"
	"log"
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

const headBlock = "head"

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

func (rpc *NodeRPC) parseResponse(resp *http.Response, checkStatusCode bool, response interface{}) (err error) {
	if resp.StatusCode > 500 {
		return NewNodeUnavailiableError(rpc.baseURL, resp.StatusCode)
	}

	if checkStatusCode && resp.StatusCode != 200 {
		return errors.Errorf("%d", resp.StatusCode)
	}
	return json.NewDecoder(resp.Body).Decode(response)
}

func (rpc *NodeRPC) getGJSONReponse(resp *http.Response, checkStatusCode bool) (result gjson.Result, err error) {
	if resp.StatusCode > 500 {
		return result, NewNodeUnavailiableError(rpc.baseURL, resp.StatusCode)
	}

	if checkStatusCode && resp.StatusCode != 200 {
		return result, errors.Errorf("%d", resp.StatusCode)
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

//nolint
func (rpc *NodeRPC) get(uri string, response interface{}) (err error) {
	url := helpers.URLJoin(rpc.baseURL, uri)
	client := http.Client{
		Timeout: rpc.timeout,
	}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return errors.Errorf("get.NewRequest: %v", err)
	}

	var resp *http.Response
	count := 0
	for ; count < rpc.retryCount; count++ {
		if resp, err = client.Do(req); err != nil {
			log.Printf("Attempt #%d: %s", count+1, err.Error())
			continue
		}
		break
	}

	if count == rpc.retryCount {
		return NewMaxRetryExceededError(rpc.baseURL)
	}
	defer resp.Body.Close()

	return rpc.parseResponse(resp, true, response)
}

//nolint
func (rpc *NodeRPC) getGJSON(uri string) (result gjson.Result, err error) {
	url := helpers.URLJoin(rpc.baseURL, uri)
	client := http.Client{
		Timeout: rpc.timeout,
	}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return result, errors.Errorf("get.NewRequest: %v", err)
	}

	var resp *http.Response
	count := 0
	for ; count < rpc.retryCount; count++ {
		if resp, err = client.Do(req); err != nil {
			log.Printf("Attempt #%d: %s", count+1, err.Error())
			continue
		}
		break
	}

	if count == rpc.retryCount {
		return result, NewMaxRetryExceededError(rpc.baseURL)
	}
	defer resp.Body.Close()

	return rpc.getGJSONReponse(resp, true)
}

//nolint
func (rpc *NodeRPC) post(uri string, data map[string]interface{}, checkStatusCode bool, response interface{}) (err error) {
	url := helpers.URLJoin(rpc.baseURL, uri)
	client := http.Client{
		Timeout: rpc.timeout,
	}

	bData, err := json.Marshal(data)
	if err != nil {
		return errors.Errorf("post.json.Marshal: %v", err)
	}

	req, err := http.NewRequest("POST", url, bytes.NewReader(bData))
	if err != nil {
		return errors.Errorf("post.NewRequest: %v", err)
	}

	var resp *http.Response
	count := 0
	for ; count < rpc.retryCount; count++ {
		if resp, err = client.Do(req); err != nil {
			log.Printf("Attempt #%d: %s", count+1, err.Error())
			continue
		}
		break
	}

	if count == rpc.retryCount {
		return NewMaxRetryExceededError(rpc.baseURL)
	}
	defer resp.Body.Close()

	return rpc.parseResponse(resp, checkStatusCode, response)
}

//nolint
func (rpc *NodeRPC) postGJSON(uri string, data map[string]interface{}, checkStatusCode bool) (result gjson.Result, err error) {
	url := helpers.URLJoin(rpc.baseURL, uri)
	client := http.Client{
		Timeout: rpc.timeout,
	}

	bData, err := json.Marshal(data)
	if err != nil {
		return result, errors.Errorf("post.json.Marshal: %v", err)
	}

	req, err := http.NewRequest("POST", url, bytes.NewReader(bData))
	if err != nil {
		return result, errors.Errorf("post.NewRequest: %v", err)
	}

	var resp *http.Response
	count := 0
	for ; count < rpc.retryCount; count++ {
		if resp, err = client.Do(req); err != nil {
			log.Printf("Attempt #%d: %s", count+1, err.Error())
			continue
		}
		break
	}

	if count == rpc.retryCount {
		return result, NewMaxRetryExceededError(rpc.baseURL)
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
	block := headBlock
	if level > 0 {
		block = fmt.Sprintf("%d", level)
	}
	err = rpc.get(fmt.Sprintf("chains/main/blocks/%s/header", block), &header)
	return
}

// GetLevelTime - get level time
func (rpc *NodeRPC) GetLevelTime(level int) (time.Time, error) {
	var head struct {
		Timestamp time.Time `json:"timestamp"`
	}
	block := headBlock
	if level > 0 {
		block = fmt.Sprintf("%d", level)
	}
	if err := rpc.get(fmt.Sprintf("chains/main/blocks/%s/header", block), &head); err != nil {
		return time.Now(), err
	}
	return head.Timestamp.UTC(), nil
}

// GetScriptJSON -
func (rpc *NodeRPC) GetScriptJSON(address string, level int64) (gjson.Result, error) {
	block := headBlock
	if level > 0 {
		block = fmt.Sprintf("%d", level)
	}

	var contract struct {
		Script stdJSON.RawMessage `json:"script"`
	}
	if err := rpc.get(fmt.Sprintf("chains/main/blocks/%s/context/contracts/%s", block, address), &contract); err != nil {
		return gjson.Result{}, err
	}

	return gjson.ParseBytes(contract.Script), nil
}

// GetScriptStorageJSON -
func (rpc *NodeRPC) GetScriptStorageJSON(address string, level int64) (gjson.Result, error) {
	block := headBlock
	if level > 0 {
		block = fmt.Sprintf("%d", level)
	}
	return rpc.getGJSON(fmt.Sprintf("chains/main/blocks/%s/context/contracts/%s/storage", block, address))
}

// GetContractBalance -
func (rpc *NodeRPC) GetContractBalance(address string, level int64) (int64, error) {
	block := headBlock
	if level > 0 {
		block = fmt.Sprintf("%d", level)
	}

	var contract struct {
		Balance int64 `json:"balance,string"`
	}
	if err := rpc.get(fmt.Sprintf("chains/main/blocks/%s/context/contracts/%s", block, address), &contract); err != nil {
		return 0, err
	}
	return contract.Balance, nil
}

// GetContractJSON -
func (rpc *NodeRPC) GetContractJSON(address string, level int64) (gjson.Result, error) {
	block := headBlock
	if level > 0 {
		block = fmt.Sprintf("%d", level)
	}

	return rpc.getGJSON(fmt.Sprintf("chains/main/blocks/%s/context/contracts/%s", block, address))
}

// GetOperations -
func (rpc *NodeRPC) GetOperations(block int64) (res gjson.Result, err error) {
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
	block := headBlock
	if level > 0 {
		block = fmt.Sprintf("%d", level)
	}
	err = rpc.get(fmt.Sprintf("chains/main/blocks/%s/context/constants", block), &constants)
	return
}

// RunCode -
func (rpc *NodeRPC) RunCode(script, storage, input gjson.Result, chainID, source, payer, entrypoint, proto string, amount, gas int64) (res gjson.Result, err error) {
	data := map[string]interface{}{
		"script":   script.Value(),
		"storage":  storage.Value(),
		"input":    input.Value(),
		"amount":   fmt.Sprintf("%d", amount),
		"chain_id": chainID,
	}

	if proto != "PsDELPH1Kxsxt8f9eWbxQeRxkjfbxoqM52jvs5Y5fBxWWh4ifpo" {
		data["balance"] = 0
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
func (rpc *NodeRPC) RunOperation(chainID, branch, source, destination string, fee, gasLimit, storageLimit, counter, amount int64, parameters gjson.Result) (res gjson.Result, err error) {
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
	block := headBlock
	if level > 0 {
		block = fmt.Sprintf("%d", level)
	}

	contract, err := rpc.getGJSON(fmt.Sprintf("chains/main/blocks/%s/context/contracts/%s/script", block, address))
	if err != nil {
		return gjson.Result{}, err
	}

	return contract.Get("code"), nil
}
