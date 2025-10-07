package noderpc

import (
	"bytes"
	"context"
	stdJSON "encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/baking-bad/bcdhub/internal/helpers"
	jsoniter "github.com/json-iterator/go"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	"golang.org/x/time/rate"
)

var json = jsoniter.ConfigCompatibleWithStandardLibrary

const (
	headBlock = "head"
	userAgent = "BetterCallDev"
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

	timeout   time.Duration
	userAgent string
	rateLimit *rate.Limiter
	needLog   bool
}

// NewNodeRPC -
func NewNodeRPC(baseURL string, opts ...NodeOption) *NodeRPC {
	node := &NodeRPC{
		baseURL:   baseURL,
		timeout:   time.Second * 10,
		userAgent: userAgent,
	}

	if bcdUserAgent := os.Getenv("BCD_USER_AGENT"); bcdUserAgent != "" {
		node.userAgent = bcdUserAgent
	}

	for _, opt := range opts {
		opt(node)
	}

	t := http.DefaultTransport.(*http.Transport).Clone()
	t.MaxIdleConns = 20
	t.MaxConnsPerHost = 20
	t.MaxIdleConnsPerHost = 20
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
		if _, err := node.GetLevel(context.Background()); err == nil {
			break
		}

		log.Warn().Msgf("Waiting node %s up 30 second...", baseURL)
		time.Sleep(time.Second * 30)
	}
	return node
}

func (rpc *NodeRPC) checkStatusCode(r io.Reader, statusCode int, checkStatusCode bool, uri string) error {
	switch {
	case statusCode == http.StatusOK:
		return nil
	case statusCode == http.StatusNotFound:
		return errors.Errorf("%s: not found", uri)
	case statusCode > http.StatusInternalServerError:
		return NewNodeUnavailiableError(rpc.baseURL, statusCode)
	case checkStatusCode:
		invalidResponseErr := newInvalidNodeResponse()
		data, err := io.ReadAll(r)
		if err != nil {
			return err
		}
		invalidResponseErr.Raw = data
		_ = json.Unmarshal(data, &invalidResponseErr.Errors)
		return invalidResponseErr
	default:
		return nil
	}
}

func (rpc *NodeRPC) parseResponse(r io.Reader, statusCode int, checkStatusCode bool, uri string, response interface{}) error {
	if err := rpc.checkStatusCode(r, statusCode, checkStatusCode, uri); err != nil {
		return fmt.Errorf("%w (%s): %w", ErrNodeRPCError, uri, err)
	}

	return json.NewDecoder(r).Decode(response)
}

func (rpc *NodeRPC) makeRequest(req *http.Request) (*http.Response, error) {
	req.Header.Set("User-Agent", rpc.userAgent)
	return rpc.client.Do(req)
}

func (rpc *NodeRPC) makeGetRequest(ctx context.Context, uri string) (*http.Response, error) {
	url, err := helpers.URLJoin(rpc.baseURL, uri)
	if err != nil {
		return nil, errors.Wrap(ErrNodeRPCError, err.Error())
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, errors.Errorf("makeGetRequest.NewRequest: %v", err)
	}
	return rpc.makeRequest(req)
}

func (rpc *NodeRPC) makePostRequest(ctx context.Context, uri string, data interface{}) (*http.Response, error) {
	url, err := helpers.URLJoin(rpc.baseURL, uri)
	if err != nil {
		return nil, errors.Wrap(ErrNodeRPCError, err.Error())
	}
	bData, err := json.Marshal(data)
	if err != nil {
		return nil, errors.Errorf("makePostRequest.json.Marshal: %v", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(bData))
	if err != nil {
		return nil, errors.Errorf("makePostRequest.NewRequest: %v", err)
	}
	return rpc.makeRequest(req)
}

func (rpc *NodeRPC) get(ctx context.Context, uri string, response interface{}) error {
	if rpc.rateLimit != nil {
		if err := rpc.rateLimit.Wait(ctx); err != nil {
			return err
		}
	}

	start := time.Now()
	defer func() {
		if rpc.needLog {
			log.Info().Str("method", "get").Int64("ms", time.Since(start).Milliseconds()).Msg(uri)
		}
	}()

	resp, err := rpc.makeGetRequest(ctx, uri)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	buffer := new(bytes.Buffer)
	if _, err = io.Copy(buffer, resp.Body); err != nil {
		return err
	}

	return rpc.parseResponse(buffer, resp.StatusCode, true, resp.Request.URL.String(), response)
}

func (rpc *NodeRPC) getRaw(ctx context.Context, uri string) ([]byte, error) {
	if rpc.rateLimit != nil {
		if err := rpc.rateLimit.Wait(ctx); err != nil {
			return nil, err
		}
	}

	start := time.Now()
	defer func() {
		if rpc.needLog {
			log.Info().Str("method", "get").Int64("ms", time.Since(start).Milliseconds()).Msg(uri)
		}
	}()

	resp, err := rpc.makeGetRequest(ctx, uri)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if err := rpc.checkStatusCode(resp.Body, resp.StatusCode, true, uri); err != nil {
		return nil, fmt.Errorf("%w (%s): %w", ErrNodeRPCError, uri, err)
	}
	return io.ReadAll(resp.Body)
}

func (rpc *NodeRPC) post(ctx context.Context, uri string, data interface{}, checkStatusCode bool, response interface{}) error {
	if rpc.rateLimit != nil {
		if err := rpc.rateLimit.Wait(ctx); err != nil {
			return err
		}
	}

	start := time.Now()
	defer func() {
		if rpc.needLog {
			log.Info().Str("method", "post").Int64("ms", time.Since(start).Milliseconds()).Msg(uri)
		}
	}()

	resp, err := rpc.makePostRequest(ctx, uri, data)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	buffer := new(bytes.Buffer)
	if _, err = io.Copy(buffer, resp.Body); err != nil {
		return err
	}

	return rpc.parseResponse(buffer, resp.StatusCode, checkStatusCode, resp.Request.URL.String(), response)
}

// Block - returns block
func (rpc *NodeRPC) Block(ctx context.Context, level int64) (block Block, err error) {
	err = rpc.get(ctx, fmt.Sprintf("chains/main/blocks/%s", getBlockString(level)), &block)
	return
}

// BlockHash - returns block's hash, its unique identifier.
func (rpc *NodeRPC) BlockHash(ctx context.Context, level int64) (hash string, err error) {
	err = rpc.get(ctx, fmt.Sprintf("chains/main/blocks/%s/hash", getBlockString(level)), &hash)
	return
}

// GetHead - get head
func (rpc *NodeRPC) GetHead(ctx context.Context) (Header, error) {
	return rpc.GetHeader(ctx, 0)
}

// GetLevel - get head level
func (rpc *NodeRPC) GetLevel(ctx context.Context) (int64, error) {
	var head struct {
		Level int64 `json:"level"`
	}
	if err := rpc.get(ctx, "chains/main/blocks/head/helpers/current_level", &head); err != nil {
		return 0, err
	}
	return head.Level, nil
}

// GetHeader - get head for certain level
func (rpc *NodeRPC) GetHeader(ctx context.Context, level int64) (header Header, err error) {
	err = rpc.get(ctx, fmt.Sprintf("chains/main/blocks/%s/header", getBlockString(level)), &header)
	return
}

// GetScriptJSON -
func (rpc *NodeRPC) GetScriptJSON(ctx context.Context, address string, level int64) (script Script, err error) {
	err = rpc.get(ctx, fmt.Sprintf("chains/main/blocks/%s/context/contracts/%s/script", getBlockString(level), address), &script)
	return
}

// GetRawScript -
func (rpc *NodeRPC) GetRawScript(ctx context.Context, address string, level int64) ([]byte, error) {
	return rpc.getRaw(ctx, fmt.Sprintf("chains/main/blocks/%s/context/contracts/%s/script", getBlockString(level), address))
}

// GetScriptStorageRaw -
func (rpc *NodeRPC) GetScriptStorageRaw(ctx context.Context, address string, level int64) ([]byte, error) {
	var response struct {
		Storage stdJSON.RawMessage `json:"storage"`
	}
	err := rpc.get(ctx, fmt.Sprintf("chains/main/blocks/%s/context/contracts/%s/script", getBlockString(level), address), &response)
	return response.Storage, err
}

// GetContractBalance -
func (rpc *NodeRPC) GetContractBalance(ctx context.Context, address string, level int64) (int64, error) {
	var balanceStr string
	if err := rpc.get(ctx, fmt.Sprintf("chains/main/blocks/%s/context/contracts/%s/balance", getBlockString(level), address), &balanceStr); err != nil {
		return 0, err
	}
	return strconv.ParseInt(balanceStr, 10, 64)
}

// GetContractData -
func (rpc *NodeRPC) GetContractData(ctx context.Context, address string, level int64) (ContractData, error) {
	var response ContractData
	err := rpc.get(ctx, fmt.Sprintf("chains/main/blocks/%s/context/contracts/%s", getBlockString(level), address), &response)
	return response, err
}

// GetOPG -
func (rpc *NodeRPC) GetOPG(ctx context.Context, block int64) (group []OperationGroup, err error) {
	err = rpc.get(ctx, fmt.Sprintf("chains/main/blocks/%s/operations/3", getBlockString(block)), &group)
	return
}

// GetLightOPG -
func (rpc *NodeRPC) GetLightOPG(ctx context.Context, block int64) (group []LightOperationGroup, err error) {
	err = rpc.get(ctx, fmt.Sprintf("chains/main/blocks/%s/operations/3", getBlockString(block)), &group)
	return
}

// GetContractsByBlock -
func (rpc *NodeRPC) GetContractsByBlock(ctx context.Context, block int64) ([]string, error) {
	if block != 1 {
		return nil, errors.Errorf("For less loading node RPC `block` value is only 1")
	}
	contracts := make([]string, 0)
	if err := rpc.get(ctx, fmt.Sprintf("chains/main/blocks/%d/context/contracts", block), &contracts); err != nil {
		return nil, err
	}
	return contracts, nil
}

// GetNetworkConstants -
func (rpc *NodeRPC) GetNetworkConstants(ctx context.Context, level int64) (constants Constants, err error) {
	err = rpc.get(ctx, fmt.Sprintf("chains/main/blocks/%s/context/constants", getBlockString(level)), &constants)
	return
}

// RunCode -
func (rpc *NodeRPC) RunCode(ctx context.Context, script, storage, input []byte, chainID, source, payer, entrypoint, proto string, amount, gas int64) (response RunCodeResponse, err error) {
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

	err = rpc.post(ctx, "chains/main/blocks/head/helpers/scripts/run_code", request, true, &response)
	return
}

// RunOperation -
func (rpc *NodeRPC) RunOperation(ctx context.Context, chainID, branch, source, destination string, fee, gasLimit, storageLimit, counter, amount int64, parameters []byte) (group OperationGroup, err error) {
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

	err = rpc.post(ctx, "chains/main/blocks/head/helpers/scripts/run_operation", request, true, &group)
	return
}

// RunOperationLight -
func (rpc *NodeRPC) RunOperationLight(ctx context.Context, chainID, branch, source, destination string, fee, gasLimit, storageLimit, counter, amount int64, parameters []byte) (group LightOperationGroup, err error) {
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

	err = rpc.post(ctx, "chains/main/blocks/head/helpers/scripts/run_operation", request, true, &group)
	return
}

// RunScriptView -
func (rpc *NodeRPC) RunScriptView(ctx context.Context, request RunScriptViewRequest) ([]byte, error) {
	var response RunScriptViewResponse
	err := rpc.post(ctx, "chains/main/blocks/head/helpers/scripts/run_script_view", request, true, &response)
	return response.Data, err
}

// GetCounter -
func (rpc *NodeRPC) GetCounter(ctx context.Context, address string) (int64, error) {
	var counter string
	if err := rpc.get(ctx, fmt.Sprintf("chains/main/blocks/head/context/contracts/%s/counter", address), &counter); err != nil {
		return 0, err
	}
	return strconv.ParseInt(counter, 10, 64)
}

// GetBigMapType -
func (rpc *NodeRPC) GetBigMapType(ctx context.Context, ptr, level int64) (bm BigMap, err error) {
	err = rpc.get(ctx, fmt.Sprintf("chains/main/blocks/%s/context/raw/json/big_maps/index/%d", getBlockString(level), ptr), &bm)
	return
}

// GetBlockMetadata -
func (rpc *NodeRPC) GetBlockMetadata(ctx context.Context, level int64) (metadata Metadata, err error) {
	err = rpc.get(ctx, fmt.Sprintf("chains/main/blocks/%s/metadata", getBlockString(level)), &metadata)
	return
}

// GetStorage -
func (rpc *NodeRPC) GetStorage(ctx context.Context, level int64, address string) (response []byte, err error) {
	response, err = rpc.getRaw(
		ctx,
		fmt.Sprintf("chains/main/blocks/%s/context/contracts/%s/storage", getBlockString(level), address),
	)
	return
}
