package noderpc

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"time"
)

// NodeRPC -
type NodeRPC struct {
	baseURL string
	network string

	outOfSyncTimeout time.Duration
	timeout          time.Duration
	retryCount       int
}

// NewNodeRPC -
func NewNodeRPC(baseURL, network string) *NodeRPC {
	return &NodeRPC{
		baseURL:          baseURL,
		network:          network,
		outOfSyncTimeout: -3 * time.Minute,
		timeout:          time.Second * 10,
		retryCount:       3,
	}
}

// SetTimeout - default is 10 sec
func (rpc *NodeRPC) SetTimeout(timeout time.Duration) {
	rpc.timeout = timeout
}

func (rpc *NodeRPC) get(uri string, ret interface{}) (err error) {
	url := fmt.Sprintf("%s/%s/%s", rpc.baseURL, rpc.network, uri)
	client := http.Client{
		Timeout: rpc.timeout,
	}

	var resp *http.Response
	count := 0
	for ; count < rpc.retryCount; count++ {
		if resp, err = client.Get(url); err != nil {
			log.Printf("Attempt #%d: %s", count+1, err.Error())
			continue
		}
		break
	}

	if count == rpc.retryCount {
		return errors.New("Max HTTP request retry exceeded")
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
func (rpc *NodeRPC) GetContract(address string) (script map[string]interface{}, err error) {
	err = rpc.get(fmt.Sprintf("chains/main/blocks/head/context/contracts/%s", address), &script)
	return
}
