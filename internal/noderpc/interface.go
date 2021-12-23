package noderpc

// INode -
type INode interface {
	GetHead() (Header, error)
	GetHeader(int64) (Header, error)
	GetScriptJSON(string, int64) (Script, error)
	GetRawScript(address string, level int64) ([]byte, error)
	GetScriptStorageRaw(string, int64) ([]byte, error)
	GetContractBalance(string, int64) (int64, error)
	GetContractData(string, int64) (ContractData, error)
	GetOPG(block int64) ([]OperationGroup, error)
	GetLightOPG(block int64) ([]LightOperationGroup, error)
	GetContractsByBlock(int64) ([]string, error)
	GetNetworkConstants(int64) (Constants, error)
	RunCode([]byte, []byte, []byte, string, string, string, string, string, int64, int64) (RunCodeResponse, error)
	RunOperation(string, string, string, string, int64, int64, int64, int64, int64, []byte) (OperationGroup, error)
	RunOperationLight(string, string, string, string, int64, int64, int64, int64, int64, []byte) (LightOperationGroup, error)
	GetCounter(string) (int64, error)
	GetBigMapType(ptr, level int64) (BigMap, error)
	GetBlockMetadata(level int64) (metadata Metadata, err error)
}
