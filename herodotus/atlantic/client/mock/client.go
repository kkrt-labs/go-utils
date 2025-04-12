// Code generated by MockGen. DO NOT EDIT.
// Source: client.go
//
// Generated by this command:
//
//	mockgen -source client.go -destination mock/client.go -package mock Client
//

// Package mock is a generated GoMock package.
package mock

import (
	context "context"
	reflect "reflect"

	atlantic "github.com/kkrt-labs/go-utils/herodotus/atlantic/client"
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

// GenerateProof mocks base method.
func (m *MockClient) GenerateProof(ctx context.Context, req *atlantic.GenerateProofRequest) (*atlantic.GenerateProofResponse, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GenerateProof", ctx, req)
	ret0, _ := ret[0].(*atlantic.GenerateProofResponse)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GenerateProof indicates an expected call of GenerateProof.
func (mr *MockClientMockRecorder) GenerateProof(ctx, req any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GenerateProof", reflect.TypeOf((*MockClient)(nil).GenerateProof), ctx, req)
}

// GetProof mocks base method.
func (m *MockClient) GetProof(ctx context.Context, atlanticQueryID string) (*atlantic.Query, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetProof", ctx, atlanticQueryID)
	ret0, _ := ret[0].(*atlantic.Query)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetProof indicates an expected call of GetProof.
func (mr *MockClientMockRecorder) GetProof(ctx, atlanticQueryID any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetProof", reflect.TypeOf((*MockClient)(nil).GetProof), ctx, atlanticQueryID)
}

// ListProofs mocks base method.
func (m *MockClient) ListProofs(ctx context.Context, req *atlantic.ListProofsRequest) (*atlantic.ListProofsResponse, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ListProofs", ctx, req)
	ret0, _ := ret[0].(*atlantic.ListProofsResponse)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ListProofs indicates an expected call of ListProofs.
func (mr *MockClientMockRecorder) ListProofs(ctx, req any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ListProofs", reflect.TypeOf((*MockClient)(nil).ListProofs), ctx, req)
}
