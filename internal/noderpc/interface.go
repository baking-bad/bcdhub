package noderpc

import "context"

// INode -
type INode interface {
	Block(context.Context, int64) (Block, error)
	GetHead(context.Context) (Header, error)
	GetHeader(context.Context, int64) (Header, error)
	GetScriptJSON(context.Context, string, int64) (Script, error)
	GetRawScript(ctx context.Context, address string, level int64) ([]byte, error)
	GetScriptStorageRaw(context.Context, string, int64) ([]byte, error)
	GetContractBalance(context.Context, string, int64) (int64, error)
	GetContractData(context.Context, string, int64) (ContractData, error)
	GetOPG(ctx context.Context, block int64) ([]OperationGroup, error)
	GetLightOPG(ctx context.Context, block int64) ([]LightOperationGroup, error)
	GetContractsByBlock(context.Context, int64) ([]string, error)
	GetNetworkConstants(context.Context, int64) (Constants, error)
	RunCode(context.Context, []byte, []byte, []byte, string, string, string, string, string, int64, int64) (RunCodeResponse, error)
	RunOperation(context.Context, string, string, string, string, int64, int64, int64, int64, int64, []byte) (OperationGroup, error)
	RunOperationLight(context.Context, string, string, string, string, int64, int64, int64, int64, int64, []byte) (LightOperationGroup, error)
	GetCounter(context.Context, string) (int64, error)
	GetBigMapType(ctx context.Context, ptr, level int64) (BigMap, error)
	GetBlockMetadata(ctx context.Context, level int64) (metadata Metadata, err error)
}
