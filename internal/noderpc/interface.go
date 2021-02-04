package noderpc

import (
	"time"

	"github.com/tidwall/gjson"
)

// INode -
type INode interface {
	GetHead() (Header, error)
	GetHeader(int64) (Header, error)
	GetLevel() (int64, error)
	GetLevelTime(int) (time.Time, error)
	GetScriptJSON(string, int64) (gjson.Result, error)
	GetScriptStorageJSON(string, int64) (gjson.Result, error)
	GetContractBalance(string, int64) (int64, error)
	GetContractData(string, int64) (ContractData, error)
	GetOperations(int64) (gjson.Result, error)
	GetContractsByBlock(int64) ([]string, error)
	GetNetworkConstants(int64) (Constants, error)
	RunCode(gjson.Result, gjson.Result, gjson.Result, string, string, string, string, string, int64, int64) (gjson.Result, error)
	RunOperation(string, string, string, string, int64, int64, int64, int64, int64, gjson.Result) (gjson.Result, error)
	GetCounter(string) (int64, error)
	GetCode(address string, level int64) (gjson.Result, error)
}
