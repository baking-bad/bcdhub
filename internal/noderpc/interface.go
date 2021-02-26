package noderpc

import (
	"time"

	"github.com/baking-bad/bcdhub/internal/bcd/ast"
)

// INode -
type INode interface {
	GetHead() (Header, error)
	GetHeader(int64) (Header, error)
	GetLevel() (int64, error)
	GetLevelTime(int) (time.Time, error)
	GetScriptJSON(string, int64) (Script, error)
	GetScriptStorageRaw(string, int64) ([]byte, error)
	GetContractBalance(string, int64) (int64, error)
	GetContractData(string, int64) (ContractData, error)
	GetOPG(block int64) ([]OperationGroup, error)
	GetContractsByBlock(int64) ([]string, error)
	GetNetworkConstants(int64) (Constants, error)
	RunCode([]byte, []byte, []byte, string, string, string, string, string, int64, int64) (RunCodeResponse, error)
	RunOperation(string, string, string, string, int64, int64, int64, int64, int64, []byte) (OperationGroup, error)
	GetCounter(string) (int64, error)
	GetCode(address string, level int64) (*ast.Script, error)
}
