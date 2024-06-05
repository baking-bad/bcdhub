package noderpc

import (
	"context"
	"crypto/rand"
	"math/big"
	"reflect"
	"time"

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

// func (p *poolItem) isBlocked() bool {
// 	return time.Now().After(p.blockTime)
// }

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
	nodes := make([]*poolItem, 0)
	for i := range p {
		nodes = append(nodes, p[i])
	}

	switch len(nodes) {
	case 0:
		return nil, errors.Errorf("No available nodes")
	case 1:
		return nodes[0], nil
	default:
		idx, err := rand.Int(rand.Reader, big.NewInt(int64(len(nodes))))
		if err != nil {
			return nil, err
		}

		return nodes[idx.Int64()], nil
	}
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

	switch len(response) {
	case 1:
		if !response[0].IsNil() {
			return reflect.Value{}, response[0].Interface().(error)
		}
		return reflect.Value{}, nil
	case 2:
		if !response[1].IsNil() {
			if IsNodeUnavailiableError(response[1].Interface().(error)) {
				node.block()
				return p.call(method, args...)
			}
			return response[0], response[1].Interface().(error)
		}
		return response[0], nil
	default:
		node.block()
		return reflect.Value{}, errors.Errorf("Invalid response length: %d", len(response))
	}
}

// Block -
func (p Pool) Block(ctx context.Context, level int64) (Block, error) {
	data, err := p.call("Block", ctx, level)
	if err != nil {
		return Block{}, err
	}
	return data.Interface().(Block), nil
}

// BlockHash -
func (p Pool) BlockHash(ctx context.Context, level int64) (string, error) {
	data, err := p.call("BlockHash", ctx, level)
	if err != nil {
		return "", err
	}
	return data.Interface().(string), nil
}

// GetHead -
func (p Pool) GetHead(ctx context.Context) (Header, error) {
	data, err := p.call("GetHead", ctx)
	if err != nil {
		return Header{}, err
	}
	return data.Interface().(Header), nil
}

// GetHeader -
func (p Pool) GetHeader(ctx context.Context, block int64) (Header, error) {
	data, err := p.call("GetHeader", ctx, block)
	if err != nil {
		return Header{}, err
	}
	return data.Interface().(Header), nil
}

// GetLevel -
func (p Pool) GetLevel(ctx context.Context) (int64, error) {
	data, err := p.call("GetLevel", ctx)
	if err != nil {
		return 0, err
	}
	return data.Int(), nil
}

// GetScriptJSON -
func (p Pool) GetScriptJSON(ctx context.Context, address string, level int64) (Script, error) {
	data, err := p.call("GetScriptJSON", ctx, address, level)
	if err != nil {
		return Script{}, err
	}
	return data.Interface().(Script), nil
}

// GetScriptStorageRaw -
func (p Pool) GetScriptStorageRaw(ctx context.Context, address string, level int64) ([]byte, error) {
	data, err := p.call("GetScriptStorageRaw", ctx, address, level)
	if err != nil {
		return nil, err
	}
	return data.Interface().([]byte), nil
}

// GetContractBalance -
func (p Pool) GetContractBalance(ctx context.Context, address string, level int64) (int64, error) {
	data, err := p.call("GetContractBalance", ctx, address, level)
	if err != nil {
		return 0, err
	}
	return data.Int(), nil
}

// GetContractData -
func (p Pool) GetContractData(ctx context.Context, address string, level int64) (ContractData, error) {
	data, err := p.call("GetContractData", ctx, address, level)
	if err != nil {
		return ContractData{}, err
	}
	return data.Interface().(ContractData), nil
}

// GetOPG -
func (p Pool) GetOPG(ctx context.Context, block int64) ([]OperationGroup, error) {
	data, err := p.call("GetOPG", ctx, block)
	if err != nil {
		return nil, err
	}
	return data.Interface().([]OperationGroup), nil
}

// GetLightOPG -
func (p Pool) GetLightOPG(ctx context.Context, block int64) ([]LightOperationGroup, error) {
	data, err := p.call("GetLightOPG", ctx, block)
	if err != nil {
		return nil, err
	}
	return data.Interface().([]LightOperationGroup), nil
}

// GetContractsByBlock -
func (p Pool) GetContractsByBlock(ctx context.Context, block int64) ([]string, error) {
	data, err := p.call("GetContractsByBlock", ctx, block)
	if err != nil {
		return nil, err
	}
	return data.Interface().([]string), nil
}

// GetNetworkConstants -
func (p Pool) GetNetworkConstants(ctx context.Context, level int64) (res Constants, err error) {
	data, err := p.call("GetNetworkConstants", ctx, level)
	if err != nil {
		return res, err
	}
	return data.Interface().(Constants), nil
}

// RunCode -
func (p Pool) RunCode(ctx context.Context, script, storage, input []byte, chainID, source, payer, entrypoint, proto string, amount, gas int64) (RunCodeResponse, error) {
	data, err := p.call("RunCode", ctx, script, storage, input, chainID, source, payer, entrypoint, proto, amount, gas)
	if err != nil {
		return RunCodeResponse{}, err
	}
	return data.Interface().(RunCodeResponse), nil
}

// RunOperation -
func (p Pool) RunOperation(ctx context.Context, chainID, branch, source, destination string, fee, gasLimit, storageLimit, counter, amount int64, parameters []byte) (OperationGroup, error) {
	data, err := p.call("RunOperation", ctx, chainID, branch, source, destination, fee, gasLimit, storageLimit, counter, amount, parameters)
	if err != nil {
		return OperationGroup{}, err
	}
	return data.Interface().(OperationGroup), nil
}

// RunOperationLight -
func (p Pool) RunOperationLight(ctx context.Context, chainID, branch, source, destination string, fee, gasLimit, storageLimit, counter, amount int64, parameters []byte) (LightOperationGroup, error) {
	data, err := p.call("RunOperationLight", ctx, chainID, branch, source, destination, fee, gasLimit, storageLimit, counter, amount, parameters)
	if err != nil {
		return LightOperationGroup{}, err
	}
	return data.Interface().(LightOperationGroup), nil
}

// GetCounter -
func (p Pool) GetCounter(ctx context.Context, address string) (int64, error) {
	data, err := p.call("GetCounter", ctx, address)
	if err != nil {
		return 0, err
	}
	return data.Int(), nil
}

// GetBigMapType -
func (p Pool) GetBigMapType(ctx context.Context, ptr, level int64) (BigMap, error) {
	data, err := p.call("GetBigMapType", ctx, ptr, level)
	if err != nil {
		return BigMap{}, err
	}
	return data.Interface().(BigMap), nil
}

// GetBlockMetadata -
func (p Pool) GetBlockMetadata(ctx context.Context, level int64) (Metadata, error) {
	data, err := p.call("GetBlockMetadata", ctx, level)
	if err != nil {
		return Metadata{}, err
	}
	return data.Interface().(Metadata), nil
}

// GetRawScript -
func (p Pool) GetRawScript(ctx context.Context, address string, level int64) ([]byte, error) {
	data, err := p.call("GetRawScript", ctx, address, level)
	if err != nil {
		return nil, err
	}
	return data.Interface().([]byte), nil
}

// RunScriptView -
func (p Pool) RunScriptView(ctx context.Context, request RunScriptViewRequest) ([]byte, error) {
	data, err := p.call("RunScriptView", ctx, request)
	if err != nil {
		return nil, err
	}
	return data.Interface().([]byte), nil
}
