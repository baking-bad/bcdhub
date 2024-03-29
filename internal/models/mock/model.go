// Code generated by MockGen. DO NOT EDIT.
// Source: model.go
//
// Generated by this command:
//
//	mockgen -source=model.go -destination=mock/model.go -package=mock -typed
//
// Package mock is a generated GoMock package.
package mock

import (
	reflect "reflect"

	gomock "go.uber.org/mock/gomock"
)

// MockModel is a mock of Model interface.
type MockModel struct {
	ctrl     *gomock.Controller
	recorder *MockModelMockRecorder
}

// MockModelMockRecorder is the mock recorder for MockModel.
type MockModelMockRecorder struct {
	mock *MockModel
}

// NewMockModel creates a new mock instance.
func NewMockModel(ctrl *gomock.Controller) *MockModel {
	mock := &MockModel{ctrl: ctrl}
	mock.recorder = &MockModelMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockModel) EXPECT() *MockModelMockRecorder {
	return m.recorder
}

// GetID mocks base method.
func (m *MockModel) GetID() int64 {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetID")
	ret0, _ := ret[0].(int64)
	return ret0
}

// GetID indicates an expected call of GetID.
func (mr *MockModelMockRecorder) GetID() *ModelGetIDCall {
	mr.mock.ctrl.T.Helper()
	call := mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetID", reflect.TypeOf((*MockModel)(nil).GetID))
	return &ModelGetIDCall{Call: call}
}

// ModelGetIDCall wrap *gomock.Call
type ModelGetIDCall struct {
	*gomock.Call
}

// Return rewrite *gomock.Call.Return
func (c *ModelGetIDCall) Return(arg0 int64) *ModelGetIDCall {
	c.Call = c.Call.Return(arg0)
	return c
}

// Do rewrite *gomock.Call.Do
func (c *ModelGetIDCall) Do(f func() int64) *ModelGetIDCall {
	c.Call = c.Call.Do(f)
	return c
}

// DoAndReturn rewrite *gomock.Call.DoAndReturn
func (c *ModelGetIDCall) DoAndReturn(f func() int64) *ModelGetIDCall {
	c.Call = c.Call.DoAndReturn(f)
	return c
}

// TableName mocks base method.
func (m *MockModel) TableName() string {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "TableName")
	ret0, _ := ret[0].(string)
	return ret0
}

// TableName indicates an expected call of TableName.
func (mr *MockModelMockRecorder) TableName() *ModelTableNameCall {
	mr.mock.ctrl.T.Helper()
	call := mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "TableName", reflect.TypeOf((*MockModel)(nil).TableName))
	return &ModelTableNameCall{Call: call}
}

// ModelTableNameCall wrap *gomock.Call
type ModelTableNameCall struct {
	*gomock.Call
}

// Return rewrite *gomock.Call.Return
func (c *ModelTableNameCall) Return(arg0 string) *ModelTableNameCall {
	c.Call = c.Call.Return(arg0)
	return c
}

// Do rewrite *gomock.Call.Do
func (c *ModelTableNameCall) Do(f func() string) *ModelTableNameCall {
	c.Call = c.Call.Do(f)
	return c
}

// DoAndReturn rewrite *gomock.Call.DoAndReturn
func (c *ModelTableNameCall) DoAndReturn(f func() string) *ModelTableNameCall {
	c.Call = c.Call.DoAndReturn(f)
	return c
}
