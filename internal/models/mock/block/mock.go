// Code generated by MockGen. DO NOT EDIT.
// Source: repository.go
//
// Generated by this command:
//
//	mockgen -source=repository.go -destination=../mock/block/mock.go -package=block -typed
//
// Package block is a generated GoMock package.
package block

import (
	context "context"
	reflect "reflect"

	block "github.com/baking-bad/bcdhub/internal/models/block"
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

// Get mocks base method.
func (m *MockRepository) Get(ctx context.Context, level int64) (block.Block, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Get", ctx, level)
	ret0, _ := ret[0].(block.Block)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Get indicates an expected call of Get.
func (mr *MockRepositoryMockRecorder) Get(ctx, level any) *RepositoryGetCall {
	mr.mock.ctrl.T.Helper()
	call := mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Get", reflect.TypeOf((*MockRepository)(nil).Get), ctx, level)
	return &RepositoryGetCall{Call: call}
}

// RepositoryGetCall wrap *gomock.Call
type RepositoryGetCall struct {
	*gomock.Call
}

// Return rewrite *gomock.Call.Return
func (c *RepositoryGetCall) Return(arg0 block.Block, arg1 error) *RepositoryGetCall {
	c.Call = c.Call.Return(arg0, arg1)
	return c
}

// Do rewrite *gomock.Call.Do
func (c *RepositoryGetCall) Do(f func(context.Context, int64) (block.Block, error)) *RepositoryGetCall {
	c.Call = c.Call.Do(f)
	return c
}

// DoAndReturn rewrite *gomock.Call.DoAndReturn
func (c *RepositoryGetCall) DoAndReturn(f func(context.Context, int64) (block.Block, error)) *RepositoryGetCall {
	c.Call = c.Call.DoAndReturn(f)
	return c
}

// Last mocks base method.
func (m *MockRepository) Last(ctx context.Context) (block.Block, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Last", ctx)
	ret0, _ := ret[0].(block.Block)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Last indicates an expected call of Last.
func (mr *MockRepositoryMockRecorder) Last(ctx any) *RepositoryLastCall {
	mr.mock.ctrl.T.Helper()
	call := mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Last", reflect.TypeOf((*MockRepository)(nil).Last), ctx)
	return &RepositoryLastCall{Call: call}
}

// RepositoryLastCall wrap *gomock.Call
type RepositoryLastCall struct {
	*gomock.Call
}

// Return rewrite *gomock.Call.Return
func (c *RepositoryLastCall) Return(arg0 block.Block, arg1 error) *RepositoryLastCall {
	c.Call = c.Call.Return(arg0, arg1)
	return c
}

// Do rewrite *gomock.Call.Do
func (c *RepositoryLastCall) Do(f func(context.Context) (block.Block, error)) *RepositoryLastCall {
	c.Call = c.Call.Do(f)
	return c
}

// DoAndReturn rewrite *gomock.Call.DoAndReturn
func (c *RepositoryLastCall) DoAndReturn(f func(context.Context) (block.Block, error)) *RepositoryLastCall {
	c.Call = c.Call.DoAndReturn(f)
	return c
}
