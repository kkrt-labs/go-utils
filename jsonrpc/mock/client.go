// Code generated by MockGen. DO NOT EDIT.
// Source: client.go
//
// Generated by this command:
//
//	mockgen -source client.go -destination mock/client.go -package jsonrpcmock Client
//

// Package jsonrpcmock is a generated GoMock package.
package jsonrpcmock

import (
	context "context"
	reflect "reflect"

	jsonrpc "github.com/kkrt-labs/go-utils/jsonrpc"
	gomock "go.uber.org/mock/gomock"
)

// MockClient is a mock of Client interface.
type MockClient struct {
	ctrl     *gomock.Controller
	recorder *MockClientMockRecorder
	isgomock struct{}
}

// MockClientMockRecorder is the mock recorder for MockClient.
type MockClientMockRecorder struct {
	mock *MockClient
}

// NewMockClient creates a new mock instance.
func NewMockClient(ctrl *gomock.Controller) *MockClient {
	mock := &MockClient{ctrl: ctrl}
	mock.recorder = &MockClientMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockClient) EXPECT() *MockClientMockRecorder {
	return m.recorder
}

// Call mocks base method.
func (m *MockClient) Call(ctx context.Context, req *jsonrpc.Request, res any) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Call", ctx, req, res)
	ret0, _ := ret[0].(error)
	return ret0
}

// Call indicates an expected call of Call.
func (mr *MockClientMockRecorder) Call(ctx, req, res any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Call", reflect.TypeOf((*MockClient)(nil).Call), ctx, req, res)
}
