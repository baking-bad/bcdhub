// Code generated by MockGen. DO NOT EDIT.
// Source: repository.go
//
// Generated by this command:
//
//	mockgen -source=repository.go -destination=../mock/contract/mock.go -package=contract -typed
//
// Package contract is a generated GoMock package.
package contract

import (
	context "context"
	reflect "reflect"

	contract "github.com/baking-bad/bcdhub/internal/models/contract"
	types "github.com/baking-bad/bcdhub/internal/models/types"
	gomock "go.uber.org/mock/gomock"
)

// MockRepository is a mock of Repository interface.
type MockRepository struct {
	ctrl     *gomock.Controller
	recorder *MockRepositoryMockRecorder
}

// MockRepositoryMockRecorder is the mock recorder for MockRepository.
type MockRepositoryMockRecorder struct {
	mock *MockRepository
}

// NewMockRepository creates a new mock instance.
func NewMockRepository(ctrl *gomock.Controller) *MockRepository {
	mock := &MockRepository{ctrl: ctrl}
	mock.recorder = &MockRepositoryMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockRepository) EXPECT() *MockRepositoryMockRecorder {
	return m.recorder
}

// AllExceptDelegators mocks base method.
func (m *MockRepository) AllExceptDelegators(ctx context.Context) ([]contract.Contract, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "AllExceptDelegators", ctx)
	ret0, _ := ret[0].([]contract.Contract)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// AllExceptDelegators indicates an expected call of AllExceptDelegators.
func (mr *MockRepositoryMockRecorder) AllExceptDelegators(ctx any) *RepositoryAllExceptDelegatorsCall {
	mr.mock.ctrl.T.Helper()
	call := mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "AllExceptDelegators", reflect.TypeOf((*MockRepository)(nil).AllExceptDelegators), ctx)
	return &RepositoryAllExceptDelegatorsCall{Call: call}
}

// RepositoryAllExceptDelegatorsCall wrap *gomock.Call
type RepositoryAllExceptDelegatorsCall struct {
	*gomock.Call
}

// Return rewrite *gomock.Call.Return
func (c *RepositoryAllExceptDelegatorsCall) Return(arg0 []contract.Contract, arg1 error) *RepositoryAllExceptDelegatorsCall {
	c.Call = c.Call.Return(arg0, arg1)
	return c
}

// Do rewrite *gomock.Call.Do
func (c *RepositoryAllExceptDelegatorsCall) Do(f func(context.Context) ([]contract.Contract, error)) *RepositoryAllExceptDelegatorsCall {
	c.Call = c.Call.Do(f)
	return c
}

// DoAndReturn rewrite *gomock.Call.DoAndReturn
func (c *RepositoryAllExceptDelegatorsCall) DoAndReturn(f func(context.Context) ([]contract.Contract, error)) *RepositoryAllExceptDelegatorsCall {
	c.Call = c.Call.DoAndReturn(f)
	return c
}

// FindOne mocks base method.
func (m *MockRepository) FindOne(ctx context.Context, tags types.Tags) (contract.Contract, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "FindOne", ctx, tags)
	ret0, _ := ret[0].(contract.Contract)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// FindOne indicates an expected call of FindOne.
func (mr *MockRepositoryMockRecorder) FindOne(ctx, tags any) *RepositoryFindOneCall {
	mr.mock.ctrl.T.Helper()
	call := mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "FindOne", reflect.TypeOf((*MockRepository)(nil).FindOne), ctx, tags)
	return &RepositoryFindOneCall{Call: call}
}

// RepositoryFindOneCall wrap *gomock.Call
type RepositoryFindOneCall struct {
	*gomock.Call
}

// Return rewrite *gomock.Call.Return
func (c *RepositoryFindOneCall) Return(arg0 contract.Contract, arg1 error) *RepositoryFindOneCall {
	c.Call = c.Call.Return(arg0, arg1)
	return c
}

// Do rewrite *gomock.Call.Do
func (c *RepositoryFindOneCall) Do(f func(context.Context, types.Tags) (contract.Contract, error)) *RepositoryFindOneCall {
	c.Call = c.Call.Do(f)
	return c
}

// DoAndReturn rewrite *gomock.Call.DoAndReturn
func (c *RepositoryFindOneCall) DoAndReturn(f func(context.Context, types.Tags) (contract.Contract, error)) *RepositoryFindOneCall {
	c.Call = c.Call.DoAndReturn(f)
	return c
}

// Get mocks base method.
func (m *MockRepository) Get(ctx context.Context, address string) (contract.Contract, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Get", ctx, address)
	ret0, _ := ret[0].(contract.Contract)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Get indicates an expected call of Get.
func (mr *MockRepositoryMockRecorder) Get(ctx, address any) *RepositoryGetCall {
	mr.mock.ctrl.T.Helper()
	call := mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Get", reflect.TypeOf((*MockRepository)(nil).Get), ctx, address)
	return &RepositoryGetCall{Call: call}
}

// RepositoryGetCall wrap *gomock.Call
type RepositoryGetCall struct {
	*gomock.Call
}

// Return rewrite *gomock.Call.Return
func (c *RepositoryGetCall) Return(arg0 contract.Contract, arg1 error) *RepositoryGetCall {
	c.Call = c.Call.Return(arg0, arg1)
	return c
}

// Do rewrite *gomock.Call.Do
func (c *RepositoryGetCall) Do(f func(context.Context, string) (contract.Contract, error)) *RepositoryGetCall {
	c.Call = c.Call.Do(f)
	return c
}

// DoAndReturn rewrite *gomock.Call.DoAndReturn
func (c *RepositoryGetCall) DoAndReturn(f func(context.Context, string) (contract.Contract, error)) *RepositoryGetCall {
	c.Call = c.Call.DoAndReturn(f)
	return c
}

// Script mocks base method.
func (m *MockRepository) Script(ctx context.Context, address, symLink string) (contract.Script, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Script", ctx, address, symLink)
	ret0, _ := ret[0].(contract.Script)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Script indicates an expected call of Script.
func (mr *MockRepositoryMockRecorder) Script(ctx, address, symLink any) *RepositoryScriptCall {
	mr.mock.ctrl.T.Helper()
	call := mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Script", reflect.TypeOf((*MockRepository)(nil).Script), ctx, address, symLink)
	return &RepositoryScriptCall{Call: call}
}

// RepositoryScriptCall wrap *gomock.Call
type RepositoryScriptCall struct {
	*gomock.Call
}

// Return rewrite *gomock.Call.Return
func (c *RepositoryScriptCall) Return(arg0 contract.Script, arg1 error) *RepositoryScriptCall {
	c.Call = c.Call.Return(arg0, arg1)
	return c
}

// Do rewrite *gomock.Call.Do
func (c *RepositoryScriptCall) Do(f func(context.Context, string, string) (contract.Script, error)) *RepositoryScriptCall {
	c.Call = c.Call.Do(f)
	return c
}

// DoAndReturn rewrite *gomock.Call.DoAndReturn
func (c *RepositoryScriptCall) DoAndReturn(f func(context.Context, string, string) (contract.Script, error)) *RepositoryScriptCall {
	c.Call = c.Call.DoAndReturn(f)
	return c
}

// ScriptPart mocks base method.
func (m *MockRepository) ScriptPart(ctx context.Context, address, symLink, part string) ([]byte, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ScriptPart", ctx, address, symLink, part)
	ret0, _ := ret[0].([]byte)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ScriptPart indicates an expected call of ScriptPart.
func (mr *MockRepositoryMockRecorder) ScriptPart(ctx, address, symLink, part any) *RepositoryScriptPartCall {
	mr.mock.ctrl.T.Helper()
	call := mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ScriptPart", reflect.TypeOf((*MockRepository)(nil).ScriptPart), ctx, address, symLink, part)
	return &RepositoryScriptPartCall{Call: call}
}

// RepositoryScriptPartCall wrap *gomock.Call
type RepositoryScriptPartCall struct {
	*gomock.Call
}

// Return rewrite *gomock.Call.Return
func (c *RepositoryScriptPartCall) Return(arg0 []byte, arg1 error) *RepositoryScriptPartCall {
	c.Call = c.Call.Return(arg0, arg1)
	return c
}

// Do rewrite *gomock.Call.Do
func (c *RepositoryScriptPartCall) Do(f func(context.Context, string, string, string) ([]byte, error)) *RepositoryScriptPartCall {
	c.Call = c.Call.Do(f)
	return c
}

// DoAndReturn rewrite *gomock.Call.DoAndReturn
func (c *RepositoryScriptPartCall) DoAndReturn(f func(context.Context, string, string, string) ([]byte, error)) *RepositoryScriptPartCall {
	c.Call = c.Call.DoAndReturn(f)
	return c
}

// MockScriptRepository is a mock of ScriptRepository interface.
type MockScriptRepository struct {
	ctrl     *gomock.Controller
	recorder *MockScriptRepositoryMockRecorder
}

// MockScriptRepositoryMockRecorder is the mock recorder for MockScriptRepository.
type MockScriptRepositoryMockRecorder struct {
	mock *MockScriptRepository
}

// NewMockScriptRepository creates a new mock instance.
func NewMockScriptRepository(ctrl *gomock.Controller) *MockScriptRepository {
	mock := &MockScriptRepository{ctrl: ctrl}
	mock.recorder = &MockScriptRepositoryMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockScriptRepository) EXPECT() *MockScriptRepositoryMockRecorder {
	return m.recorder
}

// ByHash mocks base method.
func (m *MockScriptRepository) ByHash(ctx context.Context, hash string) (contract.Script, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ByHash", ctx, hash)
	ret0, _ := ret[0].(contract.Script)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ByHash indicates an expected call of ByHash.
func (mr *MockScriptRepositoryMockRecorder) ByHash(ctx, hash any) *ScriptRepositoryByHashCall {
	mr.mock.ctrl.T.Helper()
	call := mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ByHash", reflect.TypeOf((*MockScriptRepository)(nil).ByHash), ctx, hash)
	return &ScriptRepositoryByHashCall{Call: call}
}

// ScriptRepositoryByHashCall wrap *gomock.Call
type ScriptRepositoryByHashCall struct {
	*gomock.Call
}

// Return rewrite *gomock.Call.Return
func (c *ScriptRepositoryByHashCall) Return(arg0 contract.Script, arg1 error) *ScriptRepositoryByHashCall {
	c.Call = c.Call.Return(arg0, arg1)
	return c
}

// Do rewrite *gomock.Call.Do
func (c *ScriptRepositoryByHashCall) Do(f func(context.Context, string) (contract.Script, error)) *ScriptRepositoryByHashCall {
	c.Call = c.Call.Do(f)
	return c
}

// DoAndReturn rewrite *gomock.Call.DoAndReturn
func (c *ScriptRepositoryByHashCall) DoAndReturn(f func(context.Context, string) (contract.Script, error)) *ScriptRepositoryByHashCall {
	c.Call = c.Call.DoAndReturn(f)
	return c
}

// Code mocks base method.
func (m *MockScriptRepository) Code(ctx context.Context, id int64) ([]byte, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Code", ctx, id)
	ret0, _ := ret[0].([]byte)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Code indicates an expected call of Code.
func (mr *MockScriptRepositoryMockRecorder) Code(ctx, id any) *ScriptRepositoryCodeCall {
	mr.mock.ctrl.T.Helper()
	call := mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Code", reflect.TypeOf((*MockScriptRepository)(nil).Code), ctx, id)
	return &ScriptRepositoryCodeCall{Call: call}
}

// ScriptRepositoryCodeCall wrap *gomock.Call
type ScriptRepositoryCodeCall struct {
	*gomock.Call
}

// Return rewrite *gomock.Call.Return
func (c *ScriptRepositoryCodeCall) Return(arg0 []byte, arg1 error) *ScriptRepositoryCodeCall {
	c.Call = c.Call.Return(arg0, arg1)
	return c
}

// Do rewrite *gomock.Call.Do
func (c *ScriptRepositoryCodeCall) Do(f func(context.Context, int64) ([]byte, error)) *ScriptRepositoryCodeCall {
	c.Call = c.Call.Do(f)
	return c
}

// DoAndReturn rewrite *gomock.Call.DoAndReturn
func (c *ScriptRepositoryCodeCall) DoAndReturn(f func(context.Context, int64) ([]byte, error)) *ScriptRepositoryCodeCall {
	c.Call = c.Call.DoAndReturn(f)
	return c
}

// Parameter mocks base method.
func (m *MockScriptRepository) Parameter(ctx context.Context, id int64) ([]byte, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Parameter", ctx, id)
	ret0, _ := ret[0].([]byte)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Parameter indicates an expected call of Parameter.
func (mr *MockScriptRepositoryMockRecorder) Parameter(ctx, id any) *ScriptRepositoryParameterCall {
	mr.mock.ctrl.T.Helper()
	call := mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Parameter", reflect.TypeOf((*MockScriptRepository)(nil).Parameter), ctx, id)
	return &ScriptRepositoryParameterCall{Call: call}
}

// ScriptRepositoryParameterCall wrap *gomock.Call
type ScriptRepositoryParameterCall struct {
	*gomock.Call
}

// Return rewrite *gomock.Call.Return
func (c *ScriptRepositoryParameterCall) Return(arg0 []byte, arg1 error) *ScriptRepositoryParameterCall {
	c.Call = c.Call.Return(arg0, arg1)
	return c
}

// Do rewrite *gomock.Call.Do
func (c *ScriptRepositoryParameterCall) Do(f func(context.Context, int64) ([]byte, error)) *ScriptRepositoryParameterCall {
	c.Call = c.Call.Do(f)
	return c
}

// DoAndReturn rewrite *gomock.Call.DoAndReturn
func (c *ScriptRepositoryParameterCall) DoAndReturn(f func(context.Context, int64) ([]byte, error)) *ScriptRepositoryParameterCall {
	c.Call = c.Call.DoAndReturn(f)
	return c
}

// Storage mocks base method.
func (m *MockScriptRepository) Storage(ctx context.Context, id int64) ([]byte, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Storage", ctx, id)
	ret0, _ := ret[0].([]byte)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Storage indicates an expected call of Storage.
func (mr *MockScriptRepositoryMockRecorder) Storage(ctx, id any) *ScriptRepositoryStorageCall {
	mr.mock.ctrl.T.Helper()
	call := mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Storage", reflect.TypeOf((*MockScriptRepository)(nil).Storage), ctx, id)
	return &ScriptRepositoryStorageCall{Call: call}
}

// ScriptRepositoryStorageCall wrap *gomock.Call
type ScriptRepositoryStorageCall struct {
	*gomock.Call
}

// Return rewrite *gomock.Call.Return
func (c *ScriptRepositoryStorageCall) Return(arg0 []byte, arg1 error) *ScriptRepositoryStorageCall {
	c.Call = c.Call.Return(arg0, arg1)
	return c
}

// Do rewrite *gomock.Call.Do
func (c *ScriptRepositoryStorageCall) Do(f func(context.Context, int64) ([]byte, error)) *ScriptRepositoryStorageCall {
	c.Call = c.Call.Do(f)
	return c
}

// DoAndReturn rewrite *gomock.Call.DoAndReturn
func (c *ScriptRepositoryStorageCall) DoAndReturn(f func(context.Context, int64) ([]byte, error)) *ScriptRepositoryStorageCall {
	c.Call = c.Call.DoAndReturn(f)
	return c
}

// Views mocks base method.
func (m *MockScriptRepository) Views(ctx context.Context, id int64) ([]byte, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Views", ctx, id)
	ret0, _ := ret[0].([]byte)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Views indicates an expected call of Views.
func (mr *MockScriptRepositoryMockRecorder) Views(ctx, id any) *ScriptRepositoryViewsCall {
	mr.mock.ctrl.T.Helper()
	call := mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Views", reflect.TypeOf((*MockScriptRepository)(nil).Views), ctx, id)
	return &ScriptRepositoryViewsCall{Call: call}
}

// ScriptRepositoryViewsCall wrap *gomock.Call
type ScriptRepositoryViewsCall struct {
	*gomock.Call
}

// Return rewrite *gomock.Call.Return
func (c *ScriptRepositoryViewsCall) Return(arg0 []byte, arg1 error) *ScriptRepositoryViewsCall {
	c.Call = c.Call.Return(arg0, arg1)
	return c
}

// Do rewrite *gomock.Call.Do
func (c *ScriptRepositoryViewsCall) Do(f func(context.Context, int64) ([]byte, error)) *ScriptRepositoryViewsCall {
	c.Call = c.Call.Do(f)
	return c
}

// DoAndReturn rewrite *gomock.Call.DoAndReturn
func (c *ScriptRepositoryViewsCall) DoAndReturn(f func(context.Context, int64) ([]byte, error)) *ScriptRepositoryViewsCall {
	c.Call = c.Call.DoAndReturn(f)
	return c
}

// MockConstantRepository is a mock of ConstantRepository interface.
type MockConstantRepository struct {
	ctrl     *gomock.Controller
	recorder *MockConstantRepositoryMockRecorder
}

// MockConstantRepositoryMockRecorder is the mock recorder for MockConstantRepository.
type MockConstantRepositoryMockRecorder struct {
	mock *MockConstantRepository
}

// NewMockConstantRepository creates a new mock instance.
func NewMockConstantRepository(ctrl *gomock.Controller) *MockConstantRepository {
	mock := &MockConstantRepository{ctrl: ctrl}
	mock.recorder = &MockConstantRepositoryMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockConstantRepository) EXPECT() *MockConstantRepositoryMockRecorder {
	return m.recorder
}

// All mocks base method.
func (m *MockConstantRepository) All(ctx context.Context, addresses ...string) ([]contract.GlobalConstant, error) {
	m.ctrl.T.Helper()
	varargs := []any{ctx}
	for _, a := range addresses {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "All", varargs...)
	ret0, _ := ret[0].([]contract.GlobalConstant)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// All indicates an expected call of All.
func (mr *MockConstantRepositoryMockRecorder) All(ctx any, addresses ...any) *ConstantRepositoryAllCall {
	mr.mock.ctrl.T.Helper()
	varargs := append([]any{ctx}, addresses...)
	call := mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "All", reflect.TypeOf((*MockConstantRepository)(nil).All), varargs...)
	return &ConstantRepositoryAllCall{Call: call}
}

// ConstantRepositoryAllCall wrap *gomock.Call
type ConstantRepositoryAllCall struct {
	*gomock.Call
}

// Return rewrite *gomock.Call.Return
func (c *ConstantRepositoryAllCall) Return(arg0 []contract.GlobalConstant, arg1 error) *ConstantRepositoryAllCall {
	c.Call = c.Call.Return(arg0, arg1)
	return c
}

// Do rewrite *gomock.Call.Do
func (c *ConstantRepositoryAllCall) Do(f func(context.Context, ...string) ([]contract.GlobalConstant, error)) *ConstantRepositoryAllCall {
	c.Call = c.Call.Do(f)
	return c
}

// DoAndReturn rewrite *gomock.Call.DoAndReturn
func (c *ConstantRepositoryAllCall) DoAndReturn(f func(context.Context, ...string) ([]contract.GlobalConstant, error)) *ConstantRepositoryAllCall {
	c.Call = c.Call.DoAndReturn(f)
	return c
}

// ContractList mocks base method.
func (m *MockConstantRepository) ContractList(ctx context.Context, address string, size, offset int64) ([]contract.Contract, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ContractList", ctx, address, size, offset)
	ret0, _ := ret[0].([]contract.Contract)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ContractList indicates an expected call of ContractList.
func (mr *MockConstantRepositoryMockRecorder) ContractList(ctx, address, size, offset any) *ConstantRepositoryContractListCall {
	mr.mock.ctrl.T.Helper()
	call := mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ContractList", reflect.TypeOf((*MockConstantRepository)(nil).ContractList), ctx, address, size, offset)
	return &ConstantRepositoryContractListCall{Call: call}
}

// ConstantRepositoryContractListCall wrap *gomock.Call
type ConstantRepositoryContractListCall struct {
	*gomock.Call
}

// Return rewrite *gomock.Call.Return
func (c *ConstantRepositoryContractListCall) Return(arg0 []contract.Contract, arg1 error) *ConstantRepositoryContractListCall {
	c.Call = c.Call.Return(arg0, arg1)
	return c
}

// Do rewrite *gomock.Call.Do
func (c *ConstantRepositoryContractListCall) Do(f func(context.Context, string, int64, int64) ([]contract.Contract, error)) *ConstantRepositoryContractListCall {
	c.Call = c.Call.Do(f)
	return c
}

// DoAndReturn rewrite *gomock.Call.DoAndReturn
func (c *ConstantRepositoryContractListCall) DoAndReturn(f func(context.Context, string, int64, int64) ([]contract.Contract, error)) *ConstantRepositoryContractListCall {
	c.Call = c.Call.DoAndReturn(f)
	return c
}

// ForContract mocks base method.
func (m *MockConstantRepository) ForContract(ctx context.Context, address string, size, offset int64) ([]contract.GlobalConstant, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ForContract", ctx, address, size, offset)
	ret0, _ := ret[0].([]contract.GlobalConstant)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ForContract indicates an expected call of ForContract.
func (mr *MockConstantRepositoryMockRecorder) ForContract(ctx, address, size, offset any) *ConstantRepositoryForContractCall {
	mr.mock.ctrl.T.Helper()
	call := mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ForContract", reflect.TypeOf((*MockConstantRepository)(nil).ForContract), ctx, address, size, offset)
	return &ConstantRepositoryForContractCall{Call: call}
}

// ConstantRepositoryForContractCall wrap *gomock.Call
type ConstantRepositoryForContractCall struct {
	*gomock.Call
}

// Return rewrite *gomock.Call.Return
func (c *ConstantRepositoryForContractCall) Return(arg0 []contract.GlobalConstant, arg1 error) *ConstantRepositoryForContractCall {
	c.Call = c.Call.Return(arg0, arg1)
	return c
}

// Do rewrite *gomock.Call.Do
func (c *ConstantRepositoryForContractCall) Do(f func(context.Context, string, int64, int64) ([]contract.GlobalConstant, error)) *ConstantRepositoryForContractCall {
	c.Call = c.Call.Do(f)
	return c
}

// DoAndReturn rewrite *gomock.Call.DoAndReturn
func (c *ConstantRepositoryForContractCall) DoAndReturn(f func(context.Context, string, int64, int64) ([]contract.GlobalConstant, error)) *ConstantRepositoryForContractCall {
	c.Call = c.Call.DoAndReturn(f)
	return c
}

// Get mocks base method.
func (m *MockConstantRepository) Get(ctx context.Context, address string) (contract.GlobalConstant, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Get", ctx, address)
	ret0, _ := ret[0].(contract.GlobalConstant)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Get indicates an expected call of Get.
func (mr *MockConstantRepositoryMockRecorder) Get(ctx, address any) *ConstantRepositoryGetCall {
	mr.mock.ctrl.T.Helper()
	call := mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Get", reflect.TypeOf((*MockConstantRepository)(nil).Get), ctx, address)
	return &ConstantRepositoryGetCall{Call: call}
}

// ConstantRepositoryGetCall wrap *gomock.Call
type ConstantRepositoryGetCall struct {
	*gomock.Call
}

// Return rewrite *gomock.Call.Return
func (c *ConstantRepositoryGetCall) Return(arg0 contract.GlobalConstant, arg1 error) *ConstantRepositoryGetCall {
	c.Call = c.Call.Return(arg0, arg1)
	return c
}

// Do rewrite *gomock.Call.Do
func (c *ConstantRepositoryGetCall) Do(f func(context.Context, string) (contract.GlobalConstant, error)) *ConstantRepositoryGetCall {
	c.Call = c.Call.Do(f)
	return c
}

// DoAndReturn rewrite *gomock.Call.DoAndReturn
func (c *ConstantRepositoryGetCall) DoAndReturn(f func(context.Context, string) (contract.GlobalConstant, error)) *ConstantRepositoryGetCall {
	c.Call = c.Call.DoAndReturn(f)
	return c
}

// List mocks base method.
func (m *MockConstantRepository) List(ctx context.Context, size, offset int64, orderBy, sort string) ([]contract.ListGlobalConstantItem, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "List", ctx, size, offset, orderBy, sort)
	ret0, _ := ret[0].([]contract.ListGlobalConstantItem)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// List indicates an expected call of List.
func (mr *MockConstantRepositoryMockRecorder) List(ctx, size, offset, orderBy, sort any) *ConstantRepositoryListCall {
	mr.mock.ctrl.T.Helper()
	call := mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "List", reflect.TypeOf((*MockConstantRepository)(nil).List), ctx, size, offset, orderBy, sort)
	return &ConstantRepositoryListCall{Call: call}
}

// ConstantRepositoryListCall wrap *gomock.Call
type ConstantRepositoryListCall struct {
	*gomock.Call
}

// Return rewrite *gomock.Call.Return
func (c *ConstantRepositoryListCall) Return(arg0 []contract.ListGlobalConstantItem, arg1 error) *ConstantRepositoryListCall {
	c.Call = c.Call.Return(arg0, arg1)
	return c
}

// Do rewrite *gomock.Call.Do
func (c *ConstantRepositoryListCall) Do(f func(context.Context, int64, int64, string, string) ([]contract.ListGlobalConstantItem, error)) *ConstantRepositoryListCall {
	c.Call = c.Call.Do(f)
	return c
}

// DoAndReturn rewrite *gomock.Call.DoAndReturn
func (c *ConstantRepositoryListCall) DoAndReturn(f func(context.Context, int64, int64, string, string) ([]contract.ListGlobalConstantItem, error)) *ConstantRepositoryListCall {
	c.Call = c.Call.DoAndReturn(f)
	return c
}
