package noderpc

import (
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"time"

	"github.com/baking-bad/bcdhub/internal/helpers"
	"github.com/tidwall/gjson"
)

// NodeRPC -
type NodeRPC struct {
	baseURL string

	timeout    time.Duration
	retryCount int
}

// NewNodeRPC -
func NewNodeRPC(baseURL string) *NodeRPC {
	return &NodeRPC{
		baseURL:    baseURL,
		timeout:    time.Second * 10,
		retryCount: 3,
	}
}

// SetTimeout - default is 10 sec
func (rpc *NodeRPC) SetTimeout(timeout time.Duration) {
	rpc.timeout = timeout
}

func (rpc *NodeRPC) get(uri string) (res gjson.Result, err error) {
	url := helpers.URLJoin(rpc.baseURL, uri)
	client := http.Client{
		Timeout: rpc.timeout,
	}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return res, fmt.Errorf("get.NewRequest: %v", err)
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
		return res, errors.New("Max HTTP request retry exceeded")
	}

	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return res, fmt.Errorf("get.ReadAll: %v", err)
	}

	res = gjson.ParseBytes(b)

	resp.Body.Close()
	return
}

// GetHead - get head
func (rpc *NodeRPC) GetHead() (header Header, err error) {
	data, err := rpc.get("chains/main/blocks/head/header")
	if err != nil {
		return
	}
	header.parseGJSON(data)
	return
}

// GetLevel - get head level
func (rpc *NodeRPC) GetLevel() (int64, error) {
	head, err := rpc.get("chains/main/blocks/head/header")
	if err != nil {
		return 0, err
	}
	return head.Get("level").Int(), nil
}

// GetHeader - get head
func (rpc *NodeRPC) GetHeader(level int64) (header Header, err error) {
	block := "head"
	if level > 0 {
		block = fmt.Sprintf("%d", level)
	}
	data, err := rpc.get(fmt.Sprintf("chains/main/blocks/%s/header", block))
	if err != nil {
		return
	}
	header.parseGJSON(data)
	return
}

// GetLevelTime - get level time
func (rpc *NodeRPC) GetLevelTime(level int) (time.Time, error) {
	block := "head"
	if level > 0 {
		block = fmt.Sprintf("%d", level)
	}
	head, err := rpc.get(fmt.Sprintf("chains/main/blocks/%s/header", block))
	if err != nil {
		return time.Now(), err
	}
	return head.Get("timestamp").Time().UTC(), nil
}

// GetScriptJSON -
func (rpc *NodeRPC) GetScriptJSON(address string, level int64) (gjson.Result, error) {
	block := "head"
	if level > 0 {
		block = fmt.Sprintf("%d", level)
	}

	contract, err := rpc.get(fmt.Sprintf("chains/main/blocks/%s/context/contracts/%s", block, address))
	if err != nil {
		return gjson.Result{}, err
	}

	return contract.Get("script"), nil
}

// GetScriptStorageJSON -
func (rpc *NodeRPC) GetScriptStorageJSON(address string, level int64) (gjson.Result, error) {
	block := "head"
	if level > 0 {
		block = fmt.Sprintf("%d", level)
	}

	return rpc.get(fmt.Sprintf("chains/main/blocks/%s/context/contracts/%s/storage", block, address))
}

// GetContractBalance -
func (rpc *NodeRPC) GetContractBalance(address string, level int64) (int64, error) {
	block := "head"
	if level > 0 {
		block = fmt.Sprintf("%d", level)
	}
	contract, err := rpc.get(fmt.Sprintf("chains/main/blocks/%s/context/contracts/%s", block, address))
	if err != nil {
		return 0, err
	}

	return contract.Get("balance").Int(), nil
}

// GetContractJSON -
func (rpc *NodeRPC) GetContractJSON(address string, level int64) (gjson.Result, error) {
	block := "head"
	if level > 0 {
		block = fmt.Sprintf("%d", level)
	}

	return rpc.get(fmt.Sprintf("chains/main/blocks/%s/context/contracts/%s", block, address))
}

// GetOperations -
func (rpc *NodeRPC) GetOperations(block int64) (res gjson.Result, err error) {
	return rpc.get(fmt.Sprintf("chains/main/blocks/%d/operations/3", block))
}

// GetContractsByBlock -
func (rpc *NodeRPC) GetContractsByBlock(block int64) ([]string, error) {
	if block != 1 {
		return nil, fmt.Errorf("For less loading node RPC `block` value is only 1")
	}
	data, err := rpc.get(fmt.Sprintf("chains/main/blocks/%d/context/contracts", block))
	if err != nil {
		return nil, err
	}
	contracts := make([]string, 0)
	for _, item := range data.Array() {
		contracts = append(contracts, item.String())
	}
	return contracts, nil
}

// GetNetworkConstants -
func (rpc *NodeRPC) GetNetworkConstants() (res gjson.Result, err error) {
	return rpc.get("chains/main/blocks/head/context/constants")
}
