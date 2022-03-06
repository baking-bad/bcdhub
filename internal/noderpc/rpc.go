package noderpc

import (
	"bytes"
	stdJSON "encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/baking-bad/bcdhub/internal/helpers"
	"github.com/baking-bad/bcdhub/internal/logger"
	jsoniter "github.com/json-iterator/go"
	"github.com/pkg/errors"
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
	client  *http.Client

	timeout    time.Duration
	cacheDir   string
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

	t := http.DefaultTransport.(*http.Transport).Clone()
	t.MaxIdleConns = 100
	t.MaxConnsPerHost = 100
	t.MaxIdleConnsPerHost = 100

	node.client = &http.Client{
		Timeout:   node.timeout,
		Transport: t,
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

		logger.Warning().Msgf("Waiting node %s up 30 second...", baseURL)
		time.Sleep(time.Second * 30)
	}
	return node
}

func (rpc *NodeRPC) cached() bool {
	return rpc.cacheDir != ""
}

func (rpc *NodeRPC) cachePath(uri string) string {
	return path.Join(rpc.cacheDir, uri)
}

func (rpc *NodeRPC) isHead(path string) bool {
	return strings.Contains(path, "/head/")
}

func (rpc *NodeRPC) checkStatusCode(resp *http.Response, checkStatusCode bool) error {
	switch {
	case resp.StatusCode == http.StatusOK:
		return nil
	case resp.StatusCode > http.StatusInternalServerError:
		return NewNodeUnavailiableError(rpc.baseURL, resp.StatusCode)
	case checkStatusCode:
		invalidResponseErr := newInvalidNodeResponse()
		data, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return errors.Wrap(invalidResponseErr, err.Error())
		}
		invalidResponseErr.Raw = data
		if json.Valid(data) {
			if err := json.Unmarshal(data, &invalidResponseErr.Errors); err != nil {
				return errors.Wrap(invalidResponseErr, err.Error())
			}
		}
		return invalidResponseErr
	default:
		return nil
	}
}

func (rpc *NodeRPC) parseResponse(resp *http.Response, checkStatusCode bool, uri string, response interface{}) error {
	if err := rpc.checkStatusCode(resp, checkStatusCode); err != nil {
		return errors.Wrap(err, ErrNodeRPCError)
	}

	if rpc.cached() && uri != "" && !rpc.isHead(uri) {
		cachePath := rpc.cachePath(uri)
		dirPath := path.Dir(cachePath)
		if _, err := os.Stat(dirPath); os.IsNotExist(err) {
			if err := os.MkdirAll(dirPath, 0700); err != nil {
				return err
			}
		}
		f, err := os.Create(cachePath)
		if err != nil {
			return err
		}
		defer f.Close()

		return json.NewDecoder(io.TeeReader(resp.Body, f)).Decode(response)
	}

	return json.NewDecoder(resp.Body).Decode(response)
}

func (rpc *NodeRPC) makeRequest(req *http.Request) (*http.Response, error) {
	count := 0
	for ; count < rpc.retryCount; count++ {
		resp, err := rpc.client.Do(req)
		if err != nil {
			logger.Warning().Msgf("Attempt #%d: %s", count+1, err.Error())
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

func (rpc *NodeRPC) makePostRequest(uri string, data interface{}) (*http.Response, error) {
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

func (rpc *NodeRPC) get(uri string, response interface{}) error {
	if rpc.cached() && !rpc.isHead(uri) {
		if f, err := os.Open(rpc.cachePath(uri)); err == nil {
			defer f.Close()
			if err := json.NewDecoder(f).Decode(response); err == nil {
				return nil
			}
		}
	}

	resp, err := rpc.makeGetRequest(uri)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return rpc.parseResponse(resp, true, uri, response)
}

func (rpc *NodeRPC) getRaw(uri string) ([]byte, error) {
	resp, err := rpc.makeGetRequest(uri)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if err := rpc.checkStatusCode(resp, true); err != nil {
		return nil, errors.Wrap(err, ErrNodeRPCError)
	}
	return ioutil.ReadAll(resp.Body)
}

//nolint
func (rpc *NodeRPC) post(uri string, data interface{}, checkStatusCode bool, response interface{}) error {
	resp, err := rpc.makePostRequest(uri, data)
	if err != nil {
		return NewMaxRetryExceededError(rpc.baseURL)
	}
	defer resp.Body.Close()

	return rpc.parseResponse(resp, checkStatusCode, "", response)
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

// GetScriptJSON -
func (rpc *NodeRPC) GetScriptJSON(address string, level int64) (script Script, err error) {
	err = rpc.get(fmt.Sprintf("chains/main/blocks/%s/context/contracts/%s/script", getBlockString(level), address), &script)
	return
}

// GetRawScript -
func (rpc *NodeRPC) GetRawScript(address string, level int64) ([]byte, error) {
	return rpc.getRaw(fmt.Sprintf("chains/main/blocks/%s/context/contracts/%s/script", getBlockString(level), address))
}

// GetScriptStorageRaw -
func (rpc *NodeRPC) GetScriptStorageRaw(address string, level int64) ([]byte, error) {
	var response struct {
		Storage stdJSON.RawMessage `json:"storage"`
	}
	err := rpc.get(fmt.Sprintf("chains/main/blocks/%s/context/contracts/%s/script", getBlockString(level), address), &response)
	return response.Storage, err
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

// GetOPG -
func (rpc *NodeRPC) GetOPG(block int64) (group []OperationGroup, err error) {
	err = rpc.get(fmt.Sprintf("chains/main/blocks/%s/operations/3", getBlockString(block)), &group)
	return
}

// GetLightOPG -
func (rpc *NodeRPC) GetLightOPG(block int64) (group []LightOperationGroup, err error) {
	err = rpc.get(fmt.Sprintf("chains/main/blocks/%s/operations/3", getBlockString(block)), &group)
	return
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
func (rpc *NodeRPC) RunCode(script, storage, input []byte, chainID, source, payer, entrypoint, proto string, amount, gas int64) (response RunCodeResponse, err error) {
	request := runCodeRequest{
		Script:  script,
		Storage: storage,
		Input:   input,
		Amount:  amount,
		ChainID: chainID,
	}

	if chainID != "NetXm8tYqnMWky1" {
		request.Balance = "0"
	}
	if gas != 0 {
		request.Gas = gas
	}
	if source != "" {
		request.Source = source
	}
	if payer != "" {
		request.Payer = payer
	}
	if entrypoint != "" {
		request.Entrypoint = entrypoint
	}

	err = rpc.post("chains/main/blocks/head/helpers/scripts/run_code", request, true, &response)
	return
}

// RunOperation -
func (rpc *NodeRPC) RunOperation(chainID, branch, source, destination string, fee, gasLimit, storageLimit, counter, amount int64, parameters []byte) (group OperationGroup, err error) {
	request := runOperationRequest{
		ChainID: chainID,
		Operation: runOperationItem{
			Branch:    branch,
			Signature: "sigUHx32f9wesZ1n2BWpixXz4AQaZggEtchaQNHYGRCoWNAXx45WGW2ua3apUUUAGMLPwAU41QoaFCzVSL61VaessLg4YbbP", // base58_encode(b'0' * 64, b'sig').decode()
			Contents: []runOperationItemContent{
				{
					Kind:         "transaction",
					Fee:          fee,
					Counter:      counter,
					GasLimit:     gasLimit,
					StorageLimit: storageLimit,
					Source:       source,
					Destination:  destination,
					Amount:       amount,
					Parameters:   parameters,
				},
			},
		},
	}

	err = rpc.post("chains/main/blocks/head/helpers/scripts/run_operation", request, true, &group)
	return
}

// RunOperationLight -
func (rpc *NodeRPC) RunOperationLight(chainID, branch, source, destination string, fee, gasLimit, storageLimit, counter, amount int64, parameters []byte) (group LightOperationGroup, err error) {
	request := runOperationRequest{
		ChainID: chainID,
		Operation: runOperationItem{
			Branch:    branch,
			Signature: "sigUHx32f9wesZ1n2BWpixXz4AQaZggEtchaQNHYGRCoWNAXx45WGW2ua3apUUUAGMLPwAU41QoaFCzVSL61VaessLg4YbbP", // base58_encode(b'0' * 64, b'sig').decode()
			Contents: []runOperationItemContent{
				{
					Kind:         "transaction",
					Fee:          fee,
					Counter:      counter,
					GasLimit:     gasLimit,
					StorageLimit: storageLimit,
					Source:       source,
					Destination:  destination,
					Amount:       amount,
					Parameters:   parameters,
				},
			},
		},
	}

	err = rpc.post("chains/main/blocks/head/helpers/scripts/run_operation", request, true, &group)
	return
}

// GetCounter -
func (rpc *NodeRPC) GetCounter(address string) (int64, error) {
	var counter string
	if err := rpc.get(fmt.Sprintf("chains/main/blocks/head/context/contracts/%s/counter", address), &counter); err != nil {
		return 0, err
	}
	return strconv.ParseInt(counter, 10, 64)
}

// GetBigMapType -
func (rpc *NodeRPC) GetBigMapType(ptr, level int64) (bm BigMap, err error) {
	err = rpc.get(fmt.Sprintf("chains/main/blocks/%s/context/raw/json/big_maps/index/%d", getBlockString(level), ptr), &bm)
	return
}

// GetBlockMetadata -
func (rpc *NodeRPC) GetBlockMetadata(level int64) (metadata Metadata, err error) {
	err = rpc.get(fmt.Sprintf("chains/main/blocks/%s/metadata", getBlockString(level)), &metadata)
	return
}

// RollbackCache -
func (rpc *NodeRPC) RollbackCache(level int64) error {
	if !rpc.cached() {
		return nil
	}

	dirPath := rpc.cachePath(fmt.Sprintf("chains/main/blocks/%d", level))
	if _, err := os.Stat(dirPath); os.IsNotExist(err) {
		return nil
	}
	return os.RemoveAll(dirPath)
}
