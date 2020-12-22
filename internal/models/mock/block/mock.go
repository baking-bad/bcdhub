// Code generated by MockGen. DO NOT EDIT.
// Source: block/repository.go

// Package mock_block is a generated GoMock package.
package mock_block

import (
	block "github.com/baking-bad/bcdhub/internal/models/block"
	gomock "github.com/golang/mock/gomock"
	reflect "reflect"
)

// MockRepository is a mock of Repository interface
type MockRepository struct {
	ctrl     *gomock.Controller
	recorder *MockRepositoryMockRecorder
}

// MockRepositoryMockRecorder is the mock recorder for MockRepository
type MockRepositoryMockRecorder struct {
	mock *MockRepository
}

// NewMockRepository creates a new mock instance
func NewMockRepository(ctrl *gomock.Controller) *MockRepository {
	mock := &MockRepository{ctrl: ctrl}
	mock.recorder = &MockRepositoryMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockRepository) EXPECT() *MockRepositoryMockRecorder {
	return m.recorder
}

// GetBlock mocks base method
func (m *MockRepository) GetBlock(arg0 string, arg1 int64) (block.Block, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetBlock", arg0, arg1)
	ret0, _ := ret[0].(block.Block)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetBlock indicates an expected call of GetBlock
func (mr *MockRepositoryMockRecorder) GetBlock(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetBlock", reflect.TypeOf((*MockRepository)(nil).GetBlock), arg0, arg1)
}

// GetLastBlock mocks base method
func (m *MockRepository) GetLastBlock(arg0 string) (block.Block, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetLastBlock", arg0)
	ret0, _ := ret[0].(block.Block)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetLastBlock indicates an expected call of GetLastBlock
func (mr *MockRepositoryMockRecorder) GetLastBlock(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetLastBlock", reflect.TypeOf((*MockRepository)(nil).GetLastBlock), arg0)
}

// GetLastBlocks mocks base method
func (m *MockRepository) GetLastBlocks() ([]block.Block, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetLastBlocks")
	ret0, _ := ret[0].([]block.Block)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetLastBlocks indicates an expected call of GetLastBlocks
func (mr *MockRepositoryMockRecorder) GetLastBlocks() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetLastBlocks", reflect.TypeOf((*MockRepository)(nil).GetLastBlocks))
}

// GetNetworkAlias mocks base method
func (m *MockRepository) GetNetworkAlias(chainID string) (string, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetNetworkAlias", chainID)
	ret0, _ := ret[0].(string)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetNetworkAlias indicates an expected call of GetNetworkAlias
func (mr *MockRepositoryMockRecorder) GetNetworkAlias(chainID interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetNetworkAlias", reflect.TypeOf((*MockRepository)(nil).GetNetworkAlias), chainID)
}