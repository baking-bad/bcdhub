package noderpc

import (
	"math/rand"
	"reflect"
	"time"

	"github.com/baking-bad/bcdhub/internal/bcd/ast"
	"github.com/pkg/errors"
)

// Pool - node pool
type Pool []*poolItem

type poolItem struct {
	node      *NodeRPC
	blockTime time.Time
}

func newPoolItem(url string, opts ...NodeOption) *poolItem {
	return &poolItem{
		node:      NewNodeRPC(url, opts...),
		blockTime: time.Now(),
	}
}

func newWaitPoolItem(url string, opts ...NodeOption) *poolItem {
	return &poolItem{
		node:      NewWaitNodeRPC(url, opts...),
		blockTime: time.Now(),
	}
}

func (p *poolItem) block() {
	p.blockTime = time.Now().Add(time.Minute * 5)
}

func (p *poolItem) isBlocked() bool {
	return time.Now().After(p.blockTime)
}

// NewPool - creates `Pool` struct by `urls`
func NewPool(urls []string, opts ...NodeOption) Pool {
	pool := make(Pool, len(urls))
	for i := range urls {
		pool[i] = newPoolItem(urls[i], opts...)
	}
	return pool
}

// NewWaitPool -
func NewWaitPool(urls []string, opts ...NodeOption) Pool {
	pool := make(Pool, len(urls))
	for i := range urls {
		pool[i] = newWaitPoolItem(urls[i], opts...)
	}
	return pool
}

func (p Pool) getNode() (*poolItem, error) {
	rand.Seed(time.Now().UnixNano())
	nodes := make([]*poolItem, 0)
	for i := range p {
		if p[i].isBlocked() {
			nodes = append(nodes, p[i])
		}
	}

	if len(nodes) == 0 {
		return nil, errors.Errorf("No availiable nodes")
	}

	return nodes[rand.Intn(len(nodes))], nil
}

func (p Pool) call(method string, args ...interface{}) (reflect.Value, error) {
	node, err := p.getNode()
	if err != nil {
		return reflect.Value{}, err
	}
	nodeVal := reflect.ValueOf(&node.node)
	if nodeVal.Kind() == reflect.Ptr {
		nodeVal = nodeVal.Elem()
	}

	mthd := nodeVal.MethodByName(method)
	numIn := mthd.Type().NumIn()
	if numIn != len(args) {
		return reflect.Value{}, errors.Errorf("Invalid args count: wait %d got %d", numIn, len(args))
	}

	in := make([]reflect.Value, numIn)
	for i := range args {
		in[i] = reflect.ValueOf(args[i])
	}

	response := mthd.Call(in)
	if len(response) != 2 {
		node.block()
		return reflect.Value{}, errors.Errorf("Invalid response length: %d", len(response))
	}

	if !response[1].IsNil() {
		if IsNodeUnavailiableError(response[1].Interface().(error)) {
			node.block()
			return p.call(method, args...)
		}
		return response[0], response[1].Interface().(error)
	}
	return response[0], nil
}

// GetHead -
func (p Pool) GetHead() (Header, error) {
	data, err := p.call("GetHead")
	if err != nil {
		return Header{}, err
	}
	return data.Interface().(Header), nil
}

// GetHeader -
func (p Pool) GetHeader(block int64) (Header, error) {
	data, err := p.call("GetHeader", block)
	if err != nil {
		return Header{}, err
	}
	return data.Interface().(Header), nil
}

// GetLevel -
func (p Pool) GetLevel() (int64, error) {
	data, err := p.call("GetLevel")
	if err != nil {
		return 0, err
	}
	return data.Int(), nil
}

// GetLevelTime - get level time
func (p Pool) GetLevelTime(level int) (time.Time, error) {
	data, err := p.call("GetLevelTime", level)
	if err != nil {
		return time.Now(), err
	}
	return data.Interface().(time.Time), nil
}

// GetScriptJSON -
func (p Pool) GetScriptJSON(address string, level int64) (Script, error) {
	data, err := p.call("GetScriptJSON", address, level)
	if err != nil {
		return Script{}, err
	}
	return data.Interface().(Script), nil
}

// GetScriptStorageRaw -
func (p Pool) GetScriptStorageRaw(address string, level int64) ([]byte, error) {
	data, err := p.call("GetScriptStorageRaw", address, level)
	if err != nil {
		return nil, err
	}
	return data.Interface().([]byte), nil
}

// GetContractBalance -
func (p Pool) GetContractBalance(address string, level int64) (int64, error) {
	data, err := p.call("GetContractBalance", address, level)
	if err != nil {
		return 0, err
	}
	return data.Int(), nil
}

// GetContractData -
func (p Pool) GetContractData(address string, level int64) (ContractData, error) {
	data, err := p.call("GetContractData", address, level)
	if err != nil {
		return ContractData{}, err
	}
	return data.Interface().(ContractData), nil
}

// GetOPG -
func (p Pool) GetOPG(block int64) ([]OperationGroup, error) {
	data, err := p.call("GetOPG", block)
	if err != nil {
		return nil, err
	}
	return data.Interface().([]OperationGroup), nil
}

// GetContractsByBlock -
func (p Pool) GetContractsByBlock(block int64) ([]string, error) {
	data, err := p.call("GetContractsByBlock", block)
	if err != nil {
		return nil, err
	}
	return data.Interface().([]string), nil
}

// GetNetworkConstants -
func (p Pool) GetNetworkConstants(level int64) (res Constants, err error) {
	data, err := p.call("GetNetworkConstants", level)
	if err != nil {
		return res, err
	}
	return data.Interface().(Constants), nil
}

// RunCode -
func (p Pool) RunCode(script, storage, input []byte, chainID, source, payer, entrypoint, proto string, amount, gas int64) (RunCodeResponse, error) {
	data, err := p.call("RunCode", script, storage, input, chainID, source, payer, entrypoint, proto, amount, gas)
	if err != nil {
		return RunCodeResponse{}, err
	}
	return data.Interface().(RunCodeResponse), nil
}

// RunOperation -
func (p Pool) RunOperation(chainID, branch, source, destination string, fee, gasLimit, storageLimit, counter, amount int64, parameters []byte) (OperationGroup, error) {
	data, err := p.call("RunOperation", chainID, branch, source, destination, fee, gasLimit, storageLimit, counter, amount, parameters)
	if err != nil {
		return OperationGroup{}, err
	}
	return data.Interface().(OperationGroup), nil
}

// GetCounter -
func (p Pool) GetCounter(address string) (int64, error) {
	data, err := p.call("GetCounter", address)
	if err != nil {
		return 0, err
	}
	return data.Int(), nil
}

// GetCode -
func (p Pool) GetCode(address string, level int64) (*ast.Script, error) {
	data, err := p.call("GetCode", address, level)
	if err != nil {
		return nil, err
	}
	return data.Interface().(*ast.Script), nil
}
