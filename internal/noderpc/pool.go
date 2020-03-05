package noderpc

import (
	"math/rand"
	"time"

	"github.com/tidwall/gjson"
)

// Pool - node pool
type Pool []*NodeRPC

// NewPool - creates `Pool` struct by `urls`
func NewPool(urls []string, timeout time.Duration) Pool {
	data := make(Pool, len(urls))
	for i := range urls {
		data[i] = NewNodeRPC(urls[i])
		data[i].SetTimeout(timeout)
	}
	return data
}

// GetNode - returns random node from pool
func (p Pool) GetNode() *NodeRPC {
	rand.Seed(time.Now().UnixNano())
	return p[rand.Intn(len(p))]
}

// GetLevel - get head level
func (p Pool) GetLevel() (int64, error) {
	return p.GetNode().GetLevel()
}

// GetLevelTime - get level time
func (p Pool) GetLevelTime(level int) (time.Time, error) {
	return p.GetNode().GetLevelTime(level)
}

// GetScriptJSON -
func (p Pool) GetScriptJSON(address string, level int64) (gjson.Result, error) {
	return p.GetNode().GetScriptJSON(address, level)
}

// GetContractBalance -
func (p Pool) GetContractBalance(address string, level int64) (int64, error) {
	return p.GetNode().GetContractBalance(address, level)
}

// GetContractJSON -
func (p Pool) GetContractJSON(address string, level int64) (gjson.Result, error) {
	return p.GetNode().GetContractJSON(address, level)
}

// GetOperations -
func (p Pool) GetOperations(block int64) (res gjson.Result, err error) {
	return p.GetNode().GetOperations(block)
}
