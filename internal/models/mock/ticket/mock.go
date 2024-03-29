// Code generated by MockGen. DO NOT EDIT.
// Source: repository.go
//
// Generated by this command:
//
//	mockgen -source=repository.go -destination=../mock/ticket/mock.go -package=ticket -typed
//
// Package ticket is a generated GoMock package.
package ticket

import (
	context "context"
	reflect "reflect"

	ticket "github.com/baking-bad/bcdhub/internal/models/ticket"
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

// BalancesForAccount mocks base method.
func (m *MockRepository) BalancesForAccount(ctx context.Context, accountId int64, req ticket.BalanceRequest) ([]ticket.Balance, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "BalancesForAccount", ctx, accountId, req)
	ret0, _ := ret[0].([]ticket.Balance)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// BalancesForAccount indicates an expected call of BalancesForAccount.
func (mr *MockRepositoryMockRecorder) BalancesForAccount(ctx, accountId, req any) *RepositoryBalancesForAccountCall {
	mr.mock.ctrl.T.Helper()
	call := mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "BalancesForAccount", reflect.TypeOf((*MockRepository)(nil).BalancesForAccount), ctx, accountId, req)
	return &RepositoryBalancesForAccountCall{Call: call}
}

// RepositoryBalancesForAccountCall wrap *gomock.Call
type RepositoryBalancesForAccountCall struct {
	*gomock.Call
}

// Return rewrite *gomock.Call.Return
func (c *RepositoryBalancesForAccountCall) Return(arg0 []ticket.Balance, arg1 error) *RepositoryBalancesForAccountCall {
	c.Call = c.Call.Return(arg0, arg1)
	return c
}

// Do rewrite *gomock.Call.Do
func (c *RepositoryBalancesForAccountCall) Do(f func(context.Context, int64, ticket.BalanceRequest) ([]ticket.Balance, error)) *RepositoryBalancesForAccountCall {
	c.Call = c.Call.Do(f)
	return c
}

// DoAndReturn rewrite *gomock.Call.DoAndReturn
func (c *RepositoryBalancesForAccountCall) DoAndReturn(f func(context.Context, int64, ticket.BalanceRequest) ([]ticket.Balance, error)) *RepositoryBalancesForAccountCall {
	c.Call = c.Call.DoAndReturn(f)
	return c
}

// List mocks base method.
func (m *MockRepository) List(ctx context.Context, ticketer string, limit, offset int64) ([]ticket.Ticket, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "List", ctx, ticketer, limit, offset)
	ret0, _ := ret[0].([]ticket.Ticket)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// List indicates an expected call of List.
func (mr *MockRepositoryMockRecorder) List(ctx, ticketer, limit, offset any) *RepositoryListCall {
	mr.mock.ctrl.T.Helper()
	call := mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "List", reflect.TypeOf((*MockRepository)(nil).List), ctx, ticketer, limit, offset)
	return &RepositoryListCall{Call: call}
}

// RepositoryListCall wrap *gomock.Call
type RepositoryListCall struct {
	*gomock.Call
}

// Return rewrite *gomock.Call.Return
func (c *RepositoryListCall) Return(arg0 []ticket.Ticket, arg1 error) *RepositoryListCall {
	c.Call = c.Call.Return(arg0, arg1)
	return c
}

// Do rewrite *gomock.Call.Do
func (c *RepositoryListCall) Do(f func(context.Context, string, int64, int64) ([]ticket.Ticket, error)) *RepositoryListCall {
	c.Call = c.Call.Do(f)
	return c
}

// DoAndReturn rewrite *gomock.Call.DoAndReturn
func (c *RepositoryListCall) DoAndReturn(f func(context.Context, string, int64, int64) ([]ticket.Ticket, error)) *RepositoryListCall {
	c.Call = c.Call.DoAndReturn(f)
	return c
}

// Updates mocks base method.
func (m *MockRepository) Updates(ctx context.Context, req ticket.UpdatesRequest) ([]ticket.TicketUpdate, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Updates", ctx, req)
	ret0, _ := ret[0].([]ticket.TicketUpdate)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Updates indicates an expected call of Updates.
func (mr *MockRepositoryMockRecorder) Updates(ctx, req any) *RepositoryUpdatesCall {
	mr.mock.ctrl.T.Helper()
	call := mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Updates", reflect.TypeOf((*MockRepository)(nil).Updates), ctx, req)
	return &RepositoryUpdatesCall{Call: call}
}

// RepositoryUpdatesCall wrap *gomock.Call
type RepositoryUpdatesCall struct {
	*gomock.Call
}

// Return rewrite *gomock.Call.Return
func (c *RepositoryUpdatesCall) Return(arg0 []ticket.TicketUpdate, arg1 error) *RepositoryUpdatesCall {
	c.Call = c.Call.Return(arg0, arg1)
	return c
}

// Do rewrite *gomock.Call.Do
func (c *RepositoryUpdatesCall) Do(f func(context.Context, ticket.UpdatesRequest) ([]ticket.TicketUpdate, error)) *RepositoryUpdatesCall {
	c.Call = c.Call.Do(f)
	return c
}

// DoAndReturn rewrite *gomock.Call.DoAndReturn
func (c *RepositoryUpdatesCall) DoAndReturn(f func(context.Context, ticket.UpdatesRequest) ([]ticket.TicketUpdate, error)) *RepositoryUpdatesCall {
	c.Call = c.Call.DoAndReturn(f)
	return c
}

// UpdatesForOperation mocks base method.
func (m *MockRepository) UpdatesForOperation(ctx context.Context, operationId int64) ([]ticket.TicketUpdate, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "UpdatesForOperation", ctx, operationId)
	ret0, _ := ret[0].([]ticket.TicketUpdate)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// UpdatesForOperation indicates an expected call of UpdatesForOperation.
func (mr *MockRepositoryMockRecorder) UpdatesForOperation(ctx, operationId any) *RepositoryUpdatesForOperationCall {
	mr.mock.ctrl.T.Helper()
	call := mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "UpdatesForOperation", reflect.TypeOf((*MockRepository)(nil).UpdatesForOperation), ctx, operationId)
	return &RepositoryUpdatesForOperationCall{Call: call}
}

// RepositoryUpdatesForOperationCall wrap *gomock.Call
type RepositoryUpdatesForOperationCall struct {
	*gomock.Call
}

// Return rewrite *gomock.Call.Return
func (c *RepositoryUpdatesForOperationCall) Return(arg0 []ticket.TicketUpdate, arg1 error) *RepositoryUpdatesForOperationCall {
	c.Call = c.Call.Return(arg0, arg1)
	return c
}

// Do rewrite *gomock.Call.Do
func (c *RepositoryUpdatesForOperationCall) Do(f func(context.Context, int64) ([]ticket.TicketUpdate, error)) *RepositoryUpdatesForOperationCall {
	c.Call = c.Call.Do(f)
	return c
}

// DoAndReturn rewrite *gomock.Call.DoAndReturn
func (c *RepositoryUpdatesForOperationCall) DoAndReturn(f func(context.Context, int64) ([]ticket.TicketUpdate, error)) *RepositoryUpdatesForOperationCall {
	c.Call = c.Call.DoAndReturn(f)
	return c
}
