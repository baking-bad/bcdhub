// Code generated by MockGen. DO NOT EDIT.
// Source: interface.go
//
// Generated by this command:
//
//	mockgen -source=interface.go -destination=mock.go -package=noderpc -typed
//
// Package noderpc is a generated GoMock package.
package noderpc

import (
	context "context"
	reflect "reflect"

	gomock "go.uber.org/mock/gomock"
)

// MockINode is a mock of INode interface.
type MockINode struct {
	ctrl     *gomock.Controller
	recorder *MockINodeMockRecorder
}

// MockINodeMockRecorder is the mock recorder for MockINode.
type MockINodeMockRecorder struct {
	mock *MockINode
}

// NewMockINode creates a new mock instance.
func NewMockINode(ctrl *gomock.Controller) *MockINode {
	mock := &MockINode{ctrl: ctrl}
	mock.recorder = &MockINodeMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockINode) EXPECT() *MockINodeMockRecorder {
	return m.recorder
}

// Block mocks base method.
func (m *MockINode) Block(arg0 context.Context, arg1 int64) (Block, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Block", arg0, arg1)
	ret0, _ := ret[0].(Block)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Block indicates an expected call of Block.
func (mr *MockINodeMockRecorder) Block(arg0, arg1 any) *INodeBlockCall {
	mr.mock.ctrl.T.Helper()
	call := mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Block", reflect.TypeOf((*MockINode)(nil).Block), arg0, arg1)
	return &INodeBlockCall{Call: call}
}

// INodeBlockCall wrap *gomock.Call
type INodeBlockCall struct {
	*gomock.Call
}

// Return rewrite *gomock.Call.Return
func (c *INodeBlockCall) Return(arg0 Block, arg1 error) *INodeBlockCall {
	c.Call = c.Call.Return(arg0, arg1)
	return c
}

// Do rewrite *gomock.Call.Do
func (c *INodeBlockCall) Do(f func(context.Context, int64) (Block, error)) *INodeBlockCall {
	c.Call = c.Call.Do(f)
	return c
}

// DoAndReturn rewrite *gomock.Call.DoAndReturn
func (c *INodeBlockCall) DoAndReturn(f func(context.Context, int64) (Block, error)) *INodeBlockCall {
	c.Call = c.Call.DoAndReturn(f)
	return c
}

// BlockHash mocks base method.
func (m *MockINode) BlockHash(arg0 context.Context, arg1 int64) (string, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "BlockHash", arg0, arg1)
	ret0, _ := ret[0].(string)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// BlockHash indicates an expected call of BlockHash.
func (mr *MockINodeMockRecorder) BlockHash(arg0, arg1 any) *INodeBlockHashCall {
	mr.mock.ctrl.T.Helper()
	call := mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "BlockHash", reflect.TypeOf((*MockINode)(nil).BlockHash), arg0, arg1)
	return &INodeBlockHashCall{Call: call}
}

// INodeBlockHashCall wrap *gomock.Call
type INodeBlockHashCall struct {
	*gomock.Call
}

// Return rewrite *gomock.Call.Return
func (c *INodeBlockHashCall) Return(arg0 string, arg1 error) *INodeBlockHashCall {
	c.Call = c.Call.Return(arg0, arg1)
	return c
}

// Do rewrite *gomock.Call.Do
func (c *INodeBlockHashCall) Do(f func(context.Context, int64) (string, error)) *INodeBlockHashCall {
	c.Call = c.Call.Do(f)
	return c
}

// DoAndReturn rewrite *gomock.Call.DoAndReturn
func (c *INodeBlockHashCall) DoAndReturn(f func(context.Context, int64) (string, error)) *INodeBlockHashCall {
	c.Call = c.Call.DoAndReturn(f)
	return c
}

// GetBigMapType mocks base method.
func (m *MockINode) GetBigMapType(ctx context.Context, ptr, level int64) (BigMap, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetBigMapType", ctx, ptr, level)
	ret0, _ := ret[0].(BigMap)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetBigMapType indicates an expected call of GetBigMapType.
func (mr *MockINodeMockRecorder) GetBigMapType(ctx, ptr, level any) *INodeGetBigMapTypeCall {
	mr.mock.ctrl.T.Helper()
	call := mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetBigMapType", reflect.TypeOf((*MockINode)(nil).GetBigMapType), ctx, ptr, level)
	return &INodeGetBigMapTypeCall{Call: call}
}

// INodeGetBigMapTypeCall wrap *gomock.Call
type INodeGetBigMapTypeCall struct {
	*gomock.Call
}

// Return rewrite *gomock.Call.Return
func (c *INodeGetBigMapTypeCall) Return(arg0 BigMap, arg1 error) *INodeGetBigMapTypeCall {
	c.Call = c.Call.Return(arg0, arg1)
	return c
}

// Do rewrite *gomock.Call.Do
func (c *INodeGetBigMapTypeCall) Do(f func(context.Context, int64, int64) (BigMap, error)) *INodeGetBigMapTypeCall {
	c.Call = c.Call.Do(f)
	return c
}

// DoAndReturn rewrite *gomock.Call.DoAndReturn
func (c *INodeGetBigMapTypeCall) DoAndReturn(f func(context.Context, int64, int64) (BigMap, error)) *INodeGetBigMapTypeCall {
	c.Call = c.Call.DoAndReturn(f)
	return c
}

// GetBlockMetadata mocks base method.
func (m *MockINode) GetBlockMetadata(ctx context.Context, level int64) (Metadata, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetBlockMetadata", ctx, level)
	ret0, _ := ret[0].(Metadata)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetBlockMetadata indicates an expected call of GetBlockMetadata.
func (mr *MockINodeMockRecorder) GetBlockMetadata(ctx, level any) *INodeGetBlockMetadataCall {
	mr.mock.ctrl.T.Helper()
	call := mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetBlockMetadata", reflect.TypeOf((*MockINode)(nil).GetBlockMetadata), ctx, level)
	return &INodeGetBlockMetadataCall{Call: call}
}

// INodeGetBlockMetadataCall wrap *gomock.Call
type INodeGetBlockMetadataCall struct {
	*gomock.Call
}

// Return rewrite *gomock.Call.Return
func (c *INodeGetBlockMetadataCall) Return(metadata Metadata, err error) *INodeGetBlockMetadataCall {
	c.Call = c.Call.Return(metadata, err)
	return c
}

// Do rewrite *gomock.Call.Do
func (c *INodeGetBlockMetadataCall) Do(f func(context.Context, int64) (Metadata, error)) *INodeGetBlockMetadataCall {
	c.Call = c.Call.Do(f)
	return c
}

// DoAndReturn rewrite *gomock.Call.DoAndReturn
func (c *INodeGetBlockMetadataCall) DoAndReturn(f func(context.Context, int64) (Metadata, error)) *INodeGetBlockMetadataCall {
	c.Call = c.Call.DoAndReturn(f)
	return c
}

// GetContractBalance mocks base method.
func (m *MockINode) GetContractBalance(arg0 context.Context, arg1 string, arg2 int64) (int64, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetContractBalance", arg0, arg1, arg2)
	ret0, _ := ret[0].(int64)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetContractBalance indicates an expected call of GetContractBalance.
func (mr *MockINodeMockRecorder) GetContractBalance(arg0, arg1, arg2 any) *INodeGetContractBalanceCall {
	mr.mock.ctrl.T.Helper()
	call := mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetContractBalance", reflect.TypeOf((*MockINode)(nil).GetContractBalance), arg0, arg1, arg2)
	return &INodeGetContractBalanceCall{Call: call}
}

// INodeGetContractBalanceCall wrap *gomock.Call
type INodeGetContractBalanceCall struct {
	*gomock.Call
}

// Return rewrite *gomock.Call.Return
func (c *INodeGetContractBalanceCall) Return(arg0 int64, arg1 error) *INodeGetContractBalanceCall {
	c.Call = c.Call.Return(arg0, arg1)
	return c
}

// Do rewrite *gomock.Call.Do
func (c *INodeGetContractBalanceCall) Do(f func(context.Context, string, int64) (int64, error)) *INodeGetContractBalanceCall {
	c.Call = c.Call.Do(f)
	return c
}

// DoAndReturn rewrite *gomock.Call.DoAndReturn
func (c *INodeGetContractBalanceCall) DoAndReturn(f func(context.Context, string, int64) (int64, error)) *INodeGetContractBalanceCall {
	c.Call = c.Call.DoAndReturn(f)
	return c
}

// GetContractData mocks base method.
func (m *MockINode) GetContractData(arg0 context.Context, arg1 string, arg2 int64) (ContractData, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetContractData", arg0, arg1, arg2)
	ret0, _ := ret[0].(ContractData)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetContractData indicates an expected call of GetContractData.
func (mr *MockINodeMockRecorder) GetContractData(arg0, arg1, arg2 any) *INodeGetContractDataCall {
	mr.mock.ctrl.T.Helper()
	call := mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetContractData", reflect.TypeOf((*MockINode)(nil).GetContractData), arg0, arg1, arg2)
	return &INodeGetContractDataCall{Call: call}
}

// INodeGetContractDataCall wrap *gomock.Call
type INodeGetContractDataCall struct {
	*gomock.Call
}

// Return rewrite *gomock.Call.Return
func (c *INodeGetContractDataCall) Return(arg0 ContractData, arg1 error) *INodeGetContractDataCall {
	c.Call = c.Call.Return(arg0, arg1)
	return c
}

// Do rewrite *gomock.Call.Do
func (c *INodeGetContractDataCall) Do(f func(context.Context, string, int64) (ContractData, error)) *INodeGetContractDataCall {
	c.Call = c.Call.Do(f)
	return c
}

// DoAndReturn rewrite *gomock.Call.DoAndReturn
func (c *INodeGetContractDataCall) DoAndReturn(f func(context.Context, string, int64) (ContractData, error)) *INodeGetContractDataCall {
	c.Call = c.Call.DoAndReturn(f)
	return c
}

// GetContractsByBlock mocks base method.
func (m *MockINode) GetContractsByBlock(arg0 context.Context, arg1 int64) ([]string, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetContractsByBlock", arg0, arg1)
	ret0, _ := ret[0].([]string)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetContractsByBlock indicates an expected call of GetContractsByBlock.
func (mr *MockINodeMockRecorder) GetContractsByBlock(arg0, arg1 any) *INodeGetContractsByBlockCall {
	mr.mock.ctrl.T.Helper()
	call := mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetContractsByBlock", reflect.TypeOf((*MockINode)(nil).GetContractsByBlock), arg0, arg1)
	return &INodeGetContractsByBlockCall{Call: call}
}

// INodeGetContractsByBlockCall wrap *gomock.Call
type INodeGetContractsByBlockCall struct {
	*gomock.Call
}

// Return rewrite *gomock.Call.Return
func (c *INodeGetContractsByBlockCall) Return(arg0 []string, arg1 error) *INodeGetContractsByBlockCall {
	c.Call = c.Call.Return(arg0, arg1)
	return c
}

// Do rewrite *gomock.Call.Do
func (c *INodeGetContractsByBlockCall) Do(f func(context.Context, int64) ([]string, error)) *INodeGetContractsByBlockCall {
	c.Call = c.Call.Do(f)
	return c
}

// DoAndReturn rewrite *gomock.Call.DoAndReturn
func (c *INodeGetContractsByBlockCall) DoAndReturn(f func(context.Context, int64) ([]string, error)) *INodeGetContractsByBlockCall {
	c.Call = c.Call.DoAndReturn(f)
	return c
}

// GetCounter mocks base method.
func (m *MockINode) GetCounter(arg0 context.Context, arg1 string) (int64, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetCounter", arg0, arg1)
	ret0, _ := ret[0].(int64)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetCounter indicates an expected call of GetCounter.
func (mr *MockINodeMockRecorder) GetCounter(arg0, arg1 any) *INodeGetCounterCall {
	mr.mock.ctrl.T.Helper()
	call := mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetCounter", reflect.TypeOf((*MockINode)(nil).GetCounter), arg0, arg1)
	return &INodeGetCounterCall{Call: call}
}

// INodeGetCounterCall wrap *gomock.Call
type INodeGetCounterCall struct {
	*gomock.Call
}

// Return rewrite *gomock.Call.Return
func (c *INodeGetCounterCall) Return(arg0 int64, arg1 error) *INodeGetCounterCall {
	c.Call = c.Call.Return(arg0, arg1)
	return c
}

// Do rewrite *gomock.Call.Do
func (c *INodeGetCounterCall) Do(f func(context.Context, string) (int64, error)) *INodeGetCounterCall {
	c.Call = c.Call.Do(f)
	return c
}

// DoAndReturn rewrite *gomock.Call.DoAndReturn
func (c *INodeGetCounterCall) DoAndReturn(f func(context.Context, string) (int64, error)) *INodeGetCounterCall {
	c.Call = c.Call.DoAndReturn(f)
	return c
}

// GetHead mocks base method.
func (m *MockINode) GetHead(arg0 context.Context) (Header, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetHead", arg0)
	ret0, _ := ret[0].(Header)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetHead indicates an expected call of GetHead.
func (mr *MockINodeMockRecorder) GetHead(arg0 any) *INodeGetHeadCall {
	mr.mock.ctrl.T.Helper()
	call := mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetHead", reflect.TypeOf((*MockINode)(nil).GetHead), arg0)
	return &INodeGetHeadCall{Call: call}
}

// INodeGetHeadCall wrap *gomock.Call
type INodeGetHeadCall struct {
	*gomock.Call
}

// Return rewrite *gomock.Call.Return
func (c *INodeGetHeadCall) Return(arg0 Header, arg1 error) *INodeGetHeadCall {
	c.Call = c.Call.Return(arg0, arg1)
	return c
}

// Do rewrite *gomock.Call.Do
func (c *INodeGetHeadCall) Do(f func(context.Context) (Header, error)) *INodeGetHeadCall {
	c.Call = c.Call.Do(f)
	return c
}

// DoAndReturn rewrite *gomock.Call.DoAndReturn
func (c *INodeGetHeadCall) DoAndReturn(f func(context.Context) (Header, error)) *INodeGetHeadCall {
	c.Call = c.Call.DoAndReturn(f)
	return c
}

// GetHeader mocks base method.
func (m *MockINode) GetHeader(arg0 context.Context, arg1 int64) (Header, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetHeader", arg0, arg1)
	ret0, _ := ret[0].(Header)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetHeader indicates an expected call of GetHeader.
func (mr *MockINodeMockRecorder) GetHeader(arg0, arg1 any) *INodeGetHeaderCall {
	mr.mock.ctrl.T.Helper()
	call := mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetHeader", reflect.TypeOf((*MockINode)(nil).GetHeader), arg0, arg1)
	return &INodeGetHeaderCall{Call: call}
}

// INodeGetHeaderCall wrap *gomock.Call
type INodeGetHeaderCall struct {
	*gomock.Call
}

// Return rewrite *gomock.Call.Return
func (c *INodeGetHeaderCall) Return(arg0 Header, arg1 error) *INodeGetHeaderCall {
	c.Call = c.Call.Return(arg0, arg1)
	return c
}

// Do rewrite *gomock.Call.Do
func (c *INodeGetHeaderCall) Do(f func(context.Context, int64) (Header, error)) *INodeGetHeaderCall {
	c.Call = c.Call.Do(f)
	return c
}

// DoAndReturn rewrite *gomock.Call.DoAndReturn
func (c *INodeGetHeaderCall) DoAndReturn(f func(context.Context, int64) (Header, error)) *INodeGetHeaderCall {
	c.Call = c.Call.DoAndReturn(f)
	return c
}

// GetLevel mocks base method.
func (m *MockINode) GetLevel(ctx context.Context) (int64, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetLevel", ctx)
	ret0, _ := ret[0].(int64)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetLevel indicates an expected call of GetLevel.
func (mr *MockINodeMockRecorder) GetLevel(ctx any) *INodeGetLevelCall {
	mr.mock.ctrl.T.Helper()
	call := mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetLevel", reflect.TypeOf((*MockINode)(nil).GetLevel), ctx)
	return &INodeGetLevelCall{Call: call}
}

// INodeGetLevelCall wrap *gomock.Call
type INodeGetLevelCall struct {
	*gomock.Call
}

// Return rewrite *gomock.Call.Return
func (c *INodeGetLevelCall) Return(arg0 int64, arg1 error) *INodeGetLevelCall {
	c.Call = c.Call.Return(arg0, arg1)
	return c
}

// Do rewrite *gomock.Call.Do
func (c *INodeGetLevelCall) Do(f func(context.Context) (int64, error)) *INodeGetLevelCall {
	c.Call = c.Call.Do(f)
	return c
}

// DoAndReturn rewrite *gomock.Call.DoAndReturn
func (c *INodeGetLevelCall) DoAndReturn(f func(context.Context) (int64, error)) *INodeGetLevelCall {
	c.Call = c.Call.DoAndReturn(f)
	return c
}

// GetLightOPG mocks base method.
func (m *MockINode) GetLightOPG(ctx context.Context, block int64) ([]LightOperationGroup, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetLightOPG", ctx, block)
	ret0, _ := ret[0].([]LightOperationGroup)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetLightOPG indicates an expected call of GetLightOPG.
func (mr *MockINodeMockRecorder) GetLightOPG(ctx, block any) *INodeGetLightOPGCall {
	mr.mock.ctrl.T.Helper()
	call := mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetLightOPG", reflect.TypeOf((*MockINode)(nil).GetLightOPG), ctx, block)
	return &INodeGetLightOPGCall{Call: call}
}

// INodeGetLightOPGCall wrap *gomock.Call
type INodeGetLightOPGCall struct {
	*gomock.Call
}

// Return rewrite *gomock.Call.Return
func (c *INodeGetLightOPGCall) Return(arg0 []LightOperationGroup, arg1 error) *INodeGetLightOPGCall {
	c.Call = c.Call.Return(arg0, arg1)
	return c
}

// Do rewrite *gomock.Call.Do
func (c *INodeGetLightOPGCall) Do(f func(context.Context, int64) ([]LightOperationGroup, error)) *INodeGetLightOPGCall {
	c.Call = c.Call.Do(f)
	return c
}

// DoAndReturn rewrite *gomock.Call.DoAndReturn
func (c *INodeGetLightOPGCall) DoAndReturn(f func(context.Context, int64) ([]LightOperationGroup, error)) *INodeGetLightOPGCall {
	c.Call = c.Call.DoAndReturn(f)
	return c
}

// GetNetworkConstants mocks base method.
func (m *MockINode) GetNetworkConstants(arg0 context.Context, arg1 int64) (Constants, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetNetworkConstants", arg0, arg1)
	ret0, _ := ret[0].(Constants)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetNetworkConstants indicates an expected call of GetNetworkConstants.
func (mr *MockINodeMockRecorder) GetNetworkConstants(arg0, arg1 any) *INodeGetNetworkConstantsCall {
	mr.mock.ctrl.T.Helper()
	call := mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetNetworkConstants", reflect.TypeOf((*MockINode)(nil).GetNetworkConstants), arg0, arg1)
	return &INodeGetNetworkConstantsCall{Call: call}
}

// INodeGetNetworkConstantsCall wrap *gomock.Call
type INodeGetNetworkConstantsCall struct {
	*gomock.Call
}

// Return rewrite *gomock.Call.Return
func (c *INodeGetNetworkConstantsCall) Return(arg0 Constants, arg1 error) *INodeGetNetworkConstantsCall {
	c.Call = c.Call.Return(arg0, arg1)
	return c
}

// Do rewrite *gomock.Call.Do
func (c *INodeGetNetworkConstantsCall) Do(f func(context.Context, int64) (Constants, error)) *INodeGetNetworkConstantsCall {
	c.Call = c.Call.Do(f)
	return c
}

// DoAndReturn rewrite *gomock.Call.DoAndReturn
func (c *INodeGetNetworkConstantsCall) DoAndReturn(f func(context.Context, int64) (Constants, error)) *INodeGetNetworkConstantsCall {
	c.Call = c.Call.DoAndReturn(f)
	return c
}

// GetOPG mocks base method.
func (m *MockINode) GetOPG(ctx context.Context, block int64) ([]OperationGroup, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetOPG", ctx, block)
	ret0, _ := ret[0].([]OperationGroup)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetOPG indicates an expected call of GetOPG.
func (mr *MockINodeMockRecorder) GetOPG(ctx, block any) *INodeGetOPGCall {
	mr.mock.ctrl.T.Helper()
	call := mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetOPG", reflect.TypeOf((*MockINode)(nil).GetOPG), ctx, block)
	return &INodeGetOPGCall{Call: call}
}

// INodeGetOPGCall wrap *gomock.Call
type INodeGetOPGCall struct {
	*gomock.Call
}

// Return rewrite *gomock.Call.Return
func (c *INodeGetOPGCall) Return(arg0 []OperationGroup, arg1 error) *INodeGetOPGCall {
	c.Call = c.Call.Return(arg0, arg1)
	return c
}

// Do rewrite *gomock.Call.Do
func (c *INodeGetOPGCall) Do(f func(context.Context, int64) ([]OperationGroup, error)) *INodeGetOPGCall {
	c.Call = c.Call.Do(f)
	return c
}

// DoAndReturn rewrite *gomock.Call.DoAndReturn
func (c *INodeGetOPGCall) DoAndReturn(f func(context.Context, int64) ([]OperationGroup, error)) *INodeGetOPGCall {
	c.Call = c.Call.DoAndReturn(f)
	return c
}

// GetRawScript mocks base method.
func (m *MockINode) GetRawScript(ctx context.Context, address string, level int64) ([]byte, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetRawScript", ctx, address, level)
	ret0, _ := ret[0].([]byte)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetRawScript indicates an expected call of GetRawScript.
func (mr *MockINodeMockRecorder) GetRawScript(ctx, address, level any) *INodeGetRawScriptCall {
	mr.mock.ctrl.T.Helper()
	call := mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetRawScript", reflect.TypeOf((*MockINode)(nil).GetRawScript), ctx, address, level)
	return &INodeGetRawScriptCall{Call: call}
}

// INodeGetRawScriptCall wrap *gomock.Call
type INodeGetRawScriptCall struct {
	*gomock.Call
}

// Return rewrite *gomock.Call.Return
func (c *INodeGetRawScriptCall) Return(arg0 []byte, arg1 error) *INodeGetRawScriptCall {
	c.Call = c.Call.Return(arg0, arg1)
	return c
}

// Do rewrite *gomock.Call.Do
func (c *INodeGetRawScriptCall) Do(f func(context.Context, string, int64) ([]byte, error)) *INodeGetRawScriptCall {
	c.Call = c.Call.Do(f)
	return c
}

// DoAndReturn rewrite *gomock.Call.DoAndReturn
func (c *INodeGetRawScriptCall) DoAndReturn(f func(context.Context, string, int64) ([]byte, error)) *INodeGetRawScriptCall {
	c.Call = c.Call.DoAndReturn(f)
	return c
}

// GetScriptJSON mocks base method.
func (m *MockINode) GetScriptJSON(arg0 context.Context, arg1 string, arg2 int64) (Script, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetScriptJSON", arg0, arg1, arg2)
	ret0, _ := ret[0].(Script)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetScriptJSON indicates an expected call of GetScriptJSON.
func (mr *MockINodeMockRecorder) GetScriptJSON(arg0, arg1, arg2 any) *INodeGetScriptJSONCall {
	mr.mock.ctrl.T.Helper()
	call := mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetScriptJSON", reflect.TypeOf((*MockINode)(nil).GetScriptJSON), arg0, arg1, arg2)
	return &INodeGetScriptJSONCall{Call: call}
}

// INodeGetScriptJSONCall wrap *gomock.Call
type INodeGetScriptJSONCall struct {
	*gomock.Call
}

// Return rewrite *gomock.Call.Return
func (c *INodeGetScriptJSONCall) Return(arg0 Script, arg1 error) *INodeGetScriptJSONCall {
	c.Call = c.Call.Return(arg0, arg1)
	return c
}

// Do rewrite *gomock.Call.Do
func (c *INodeGetScriptJSONCall) Do(f func(context.Context, string, int64) (Script, error)) *INodeGetScriptJSONCall {
	c.Call = c.Call.Do(f)
	return c
}

// DoAndReturn rewrite *gomock.Call.DoAndReturn
func (c *INodeGetScriptJSONCall) DoAndReturn(f func(context.Context, string, int64) (Script, error)) *INodeGetScriptJSONCall {
	c.Call = c.Call.DoAndReturn(f)
	return c
}

// GetScriptStorageRaw mocks base method.
func (m *MockINode) GetScriptStorageRaw(arg0 context.Context, arg1 string, arg2 int64) ([]byte, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetScriptStorageRaw", arg0, arg1, arg2)
	ret0, _ := ret[0].([]byte)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetScriptStorageRaw indicates an expected call of GetScriptStorageRaw.
func (mr *MockINodeMockRecorder) GetScriptStorageRaw(arg0, arg1, arg2 any) *INodeGetScriptStorageRawCall {
	mr.mock.ctrl.T.Helper()
	call := mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetScriptStorageRaw", reflect.TypeOf((*MockINode)(nil).GetScriptStorageRaw), arg0, arg1, arg2)
	return &INodeGetScriptStorageRawCall{Call: call}
}

// INodeGetScriptStorageRawCall wrap *gomock.Call
type INodeGetScriptStorageRawCall struct {
	*gomock.Call
}

// Return rewrite *gomock.Call.Return
func (c *INodeGetScriptStorageRawCall) Return(arg0 []byte, arg1 error) *INodeGetScriptStorageRawCall {
	c.Call = c.Call.Return(arg0, arg1)
	return c
}

// Do rewrite *gomock.Call.Do
func (c *INodeGetScriptStorageRawCall) Do(f func(context.Context, string, int64) ([]byte, error)) *INodeGetScriptStorageRawCall {
	c.Call = c.Call.Do(f)
	return c
}

// DoAndReturn rewrite *gomock.Call.DoAndReturn
func (c *INodeGetScriptStorageRawCall) DoAndReturn(f func(context.Context, string, int64) ([]byte, error)) *INodeGetScriptStorageRawCall {
	c.Call = c.Call.DoAndReturn(f)
	return c
}

// RunCode mocks base method.
func (m *MockINode) RunCode(arg0 context.Context, arg1, arg2, arg3 []byte, arg4, arg5, arg6, arg7, arg8 string, arg9, arg10 int64) (RunCodeResponse, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "RunCode", arg0, arg1, arg2, arg3, arg4, arg5, arg6, arg7, arg8, arg9, arg10)
	ret0, _ := ret[0].(RunCodeResponse)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// RunCode indicates an expected call of RunCode.
func (mr *MockINodeMockRecorder) RunCode(arg0, arg1, arg2, arg3, arg4, arg5, arg6, arg7, arg8, arg9, arg10 any) *INodeRunCodeCall {
	mr.mock.ctrl.T.Helper()
	call := mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "RunCode", reflect.TypeOf((*MockINode)(nil).RunCode), arg0, arg1, arg2, arg3, arg4, arg5, arg6, arg7, arg8, arg9, arg10)
	return &INodeRunCodeCall{Call: call}
}

// INodeRunCodeCall wrap *gomock.Call
type INodeRunCodeCall struct {
	*gomock.Call
}

// Return rewrite *gomock.Call.Return
func (c *INodeRunCodeCall) Return(arg0 RunCodeResponse, arg1 error) *INodeRunCodeCall {
	c.Call = c.Call.Return(arg0, arg1)
	return c
}

// Do rewrite *gomock.Call.Do
func (c *INodeRunCodeCall) Do(f func(context.Context, []byte, []byte, []byte, string, string, string, string, string, int64, int64) (RunCodeResponse, error)) *INodeRunCodeCall {
	c.Call = c.Call.Do(f)
	return c
}

// DoAndReturn rewrite *gomock.Call.DoAndReturn
func (c *INodeRunCodeCall) DoAndReturn(f func(context.Context, []byte, []byte, []byte, string, string, string, string, string, int64, int64) (RunCodeResponse, error)) *INodeRunCodeCall {
	c.Call = c.Call.DoAndReturn(f)
	return c
}

// RunOperation mocks base method.
func (m *MockINode) RunOperation(arg0 context.Context, arg1, arg2, arg3, arg4 string, arg5, arg6, arg7, arg8, arg9 int64, arg10 []byte) (OperationGroup, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "RunOperation", arg0, arg1, arg2, arg3, arg4, arg5, arg6, arg7, arg8, arg9, arg10)
	ret0, _ := ret[0].(OperationGroup)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// RunOperation indicates an expected call of RunOperation.
func (mr *MockINodeMockRecorder) RunOperation(arg0, arg1, arg2, arg3, arg4, arg5, arg6, arg7, arg8, arg9, arg10 any) *INodeRunOperationCall {
	mr.mock.ctrl.T.Helper()
	call := mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "RunOperation", reflect.TypeOf((*MockINode)(nil).RunOperation), arg0, arg1, arg2, arg3, arg4, arg5, arg6, arg7, arg8, arg9, arg10)
	return &INodeRunOperationCall{Call: call}
}

// INodeRunOperationCall wrap *gomock.Call
type INodeRunOperationCall struct {
	*gomock.Call
}

// Return rewrite *gomock.Call.Return
func (c *INodeRunOperationCall) Return(arg0 OperationGroup, arg1 error) *INodeRunOperationCall {
	c.Call = c.Call.Return(arg0, arg1)
	return c
}

// Do rewrite *gomock.Call.Do
func (c *INodeRunOperationCall) Do(f func(context.Context, string, string, string, string, int64, int64, int64, int64, int64, []byte) (OperationGroup, error)) *INodeRunOperationCall {
	c.Call = c.Call.Do(f)
	return c
}

// DoAndReturn rewrite *gomock.Call.DoAndReturn
func (c *INodeRunOperationCall) DoAndReturn(f func(context.Context, string, string, string, string, int64, int64, int64, int64, int64, []byte) (OperationGroup, error)) *INodeRunOperationCall {
	c.Call = c.Call.DoAndReturn(f)
	return c
}

// RunOperationLight mocks base method.
func (m *MockINode) RunOperationLight(arg0 context.Context, arg1, arg2, arg3, arg4 string, arg5, arg6, arg7, arg8, arg9 int64, arg10 []byte) (LightOperationGroup, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "RunOperationLight", arg0, arg1, arg2, arg3, arg4, arg5, arg6, arg7, arg8, arg9, arg10)
	ret0, _ := ret[0].(LightOperationGroup)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// RunOperationLight indicates an expected call of RunOperationLight.
func (mr *MockINodeMockRecorder) RunOperationLight(arg0, arg1, arg2, arg3, arg4, arg5, arg6, arg7, arg8, arg9, arg10 any) *INodeRunOperationLightCall {
	mr.mock.ctrl.T.Helper()
	call := mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "RunOperationLight", reflect.TypeOf((*MockINode)(nil).RunOperationLight), arg0, arg1, arg2, arg3, arg4, arg5, arg6, arg7, arg8, arg9, arg10)
	return &INodeRunOperationLightCall{Call: call}
}

// INodeRunOperationLightCall wrap *gomock.Call
type INodeRunOperationLightCall struct {
	*gomock.Call
}

// Return rewrite *gomock.Call.Return
func (c *INodeRunOperationLightCall) Return(arg0 LightOperationGroup, arg1 error) *INodeRunOperationLightCall {
	c.Call = c.Call.Return(arg0, arg1)
	return c
}

// Do rewrite *gomock.Call.Do
func (c *INodeRunOperationLightCall) Do(f func(context.Context, string, string, string, string, int64, int64, int64, int64, int64, []byte) (LightOperationGroup, error)) *INodeRunOperationLightCall {
	c.Call = c.Call.Do(f)
	return c
}

// DoAndReturn rewrite *gomock.Call.DoAndReturn
func (c *INodeRunOperationLightCall) DoAndReturn(f func(context.Context, string, string, string, string, int64, int64, int64, int64, int64, []byte) (LightOperationGroup, error)) *INodeRunOperationLightCall {
	c.Call = c.Call.DoAndReturn(f)
	return c
}

// RunScriptView mocks base method.
func (m *MockINode) RunScriptView(ctx context.Context, request RunScriptViewRequest) ([]byte, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "RunScriptView", ctx, request)
	ret0, _ := ret[0].([]byte)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// RunScriptView indicates an expected call of RunScriptView.
func (mr *MockINodeMockRecorder) RunScriptView(ctx, request any) *INodeRunScriptViewCall {
	mr.mock.ctrl.T.Helper()
	call := mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "RunScriptView", reflect.TypeOf((*MockINode)(nil).RunScriptView), ctx, request)
	return &INodeRunScriptViewCall{Call: call}
}

// INodeRunScriptViewCall wrap *gomock.Call
type INodeRunScriptViewCall struct {
	*gomock.Call
}

// Return rewrite *gomock.Call.Return
func (c *INodeRunScriptViewCall) Return(arg0 []byte, arg1 error) *INodeRunScriptViewCall {
	c.Call = c.Call.Return(arg0, arg1)
	return c
}

// Do rewrite *gomock.Call.Do
func (c *INodeRunScriptViewCall) Do(f func(context.Context, RunScriptViewRequest) ([]byte, error)) *INodeRunScriptViewCall {
	c.Call = c.Call.Do(f)
	return c
}

// DoAndReturn rewrite *gomock.Call.DoAndReturn
func (c *INodeRunScriptViewCall) DoAndReturn(f func(context.Context, RunScriptViewRequest) ([]byte, error)) *INodeRunScriptViewCall {
	c.Call = c.Call.DoAndReturn(f)
	return c
}
