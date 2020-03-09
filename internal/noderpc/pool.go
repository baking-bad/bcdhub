package noderpc

import (
	"fmt"
	"math/rand"
	"reflect"
	"time"

	"github.com/tidwall/gjson"
)

// Pool - node pool
type Pool []poolItem

type poolItem struct {
	node      *NodeRPC
	blockTime time.Time
}

var blockDuration = time.Minute

// NewPool - creates `Pool` struct by `urls`
func NewPool(urls []string, timeout time.Duration) Pool {
	data := make(Pool, len(urls))
	for i := range urls {
		data[i] = poolItem{
			node:      NewNodeRPC(urls[i]),
			blockTime: time.Now(),
		}
		data[i].node.SetTimeout(timeout)
	}
	return data
}

func (p Pool) getNode() poolItem {
	rand.Seed(time.Now().UnixNano())
	nodes := make([]poolItem, 0)
	for i := range p {
		if time.Now().After(p[i].blockTime) {
			nodes = append(nodes, p[i])
		}
	}

	return nodes[rand.Intn(len(nodes))]
}

func (p Pool) call(method string, args ...interface{}) (reflect.Value, error) {
	node := p.getNode()
	nodeVal := reflect.ValueOf(&node.node)
	if nodeVal.Kind() == reflect.Ptr {
		nodeVal = nodeVal.Elem()
	}

	mthd := nodeVal.MethodByName(method)
	numIn := mthd.Type().NumIn()
	if numIn != len(args) {
		return reflect.Value{}, fmt.Errorf("Invalid args count: wait %d got %d", numIn, len(args))
	}

	in := make([]reflect.Value, numIn)
	for i := range args {
		in[i] = reflect.ValueOf(args[i])
	}

	response := mthd.Call(in)
	if len(response) != 2 {
		node.blockTime = time.Now().Add(blockDuration)
		return reflect.Value{}, fmt.Errorf("Invalid response length: %d", len(response))
	}

	if !response[1].IsNil() {
		node.blockTime = time.Now().Add(blockDuration)
		return reflect.Value{}, response[1].Interface().(error)
	}
	return response[0], nil
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
func (p Pool) GetScriptJSON(address string, level int64) (gjson.Result, error) {
	data, err := p.call("GetScriptJSON", address, level)
	if err != nil {
		return gjson.Result{}, err
	}
	return data.Interface().(gjson.Result), nil
}

// GetContractBalance -
func (p Pool) GetContractBalance(address string, level int64) (int64, error) {
	data, err := p.call("GetContractBalance", address, level)
	if err != nil {
		return 0, err
	}
	return data.Int(), nil
}

// GetContractJSON -
func (p Pool) GetContractJSON(address string, level int64) (gjson.Result, error) {
	data, err := p.call("GetContractJSON", address, level)
	if err != nil {
		return gjson.Result{}, err
	}
	return data.Interface().(gjson.Result), nil
}

// GetOperations -
func (p Pool) GetOperations(block int64) (res gjson.Result, err error) {
	data, err := p.call("GetOperations", block)
	if err != nil {
		return
	}
	return data.Interface().(gjson.Result), nil
}
