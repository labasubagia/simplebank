// Code generated by MockGen. DO NOT EDIT.
// Source: distributor.go

// Package mock_worker is a generated GoMock package.
package mock_worker

import (
	context "context"
	reflect "reflect"

	asynq "github.com/hibiken/asynq"
	worker "github.com/labasubagia/simplebank/worker"
	gomock "go.uber.org/mock/gomock"
)

// MockTaskDistributor is a mock of TaskDistributor interface.
type MockTaskDistributor struct {
	ctrl     *gomock.Controller
	recorder *MockTaskDistributorMockRecorder
}

// MockTaskDistributorMockRecorder is the mock recorder for MockTaskDistributor.
type MockTaskDistributorMockRecorder struct {
	mock *MockTaskDistributor
}

// NewMockTaskDistributor creates a new mock instance.
func NewMockTaskDistributor(ctrl *gomock.Controller) *MockTaskDistributor {
	mock := &MockTaskDistributor{ctrl: ctrl}
	mock.recorder = &MockTaskDistributorMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockTaskDistributor) EXPECT() *MockTaskDistributorMockRecorder {
	return m.recorder
}

// DistributeTaskVerifyEmail mocks base method.
func (m *MockTaskDistributor) DistributeTaskVerifyEmail(ctx context.Context, payload *worker.PayloadSendVerifyEmail, opts ...asynq.Option) error {
	m.ctrl.T.Helper()
	varargs := []interface{}{ctx, payload}
	for _, a := range opts {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "DistributeTaskVerifyEmail", varargs...)
	ret0, _ := ret[0].(error)
	return ret0
}

// DistributeTaskVerifyEmail indicates an expected call of DistributeTaskVerifyEmail.
func (mr *MockTaskDistributorMockRecorder) DistributeTaskVerifyEmail(ctx, payload interface{}, opts ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{ctx, payload}, opts...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DistributeTaskVerifyEmail", reflect.TypeOf((*MockTaskDistributor)(nil).DistributeTaskVerifyEmail), varargs...)
}
