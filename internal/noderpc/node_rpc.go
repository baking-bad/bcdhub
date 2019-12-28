package noderpc

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"
)

// NodeRPC -
type NodeRPC struct {
	baseURL string
	network string

	outOfSyncTimeout time.Duration
	timeout          time.Duration
}

// NewNodeRPC -
func NewNodeRPC(baseURL, network string) *NodeRPC {
	return &NodeRPC{
		baseURL:          baseURL,
		network:          network,
		outOfSyncTimeout: -3 * time.Minute,
		timeout:          time.Second * 10,
	}
}

// SetTimeout - default is 10 sec
func (rpc *NodeRPC) SetTimeout(timeout time.Duration) {
	rpc.timeout = timeout
}

func (rpc *NodeRPC) get(uri string, ret interface{}) error {
	url := fmt.Sprintf("%s/%s/%s", rpc.baseURL, rpc.network, uri)
	client := http.Client{
		Timeout: rpc.timeout,
	}

	resp, err := client.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	err = json.Unmarshal(body, &ret)
	if err != nil {
		return err
	}
	return nil
}

// GetHead -
func (rpc *NodeRPC) GetHead() (Header, error) {
	var ret Header
	if err := rpc.get("chains/main/blocks/head/header", &ret); err != nil {
		return ret, err
	}
	return ret, nil
}

// GetContractScript -
func (rpc *NodeRPC) GetContractScript(address string) (script map[string]interface{}, err error) {
	err = rpc.get(fmt.Sprintf("chains/main/blocks/head/context/contracts/%s/script", address), &script)
	return
}
