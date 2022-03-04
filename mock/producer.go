// Code generated by MockGen. DO NOT EDIT.
// Source: github.com/conduitio/conduit-connector-kafka (interfaces: Producer)

// Package mock is a generated GoMock package.
package mock

import (
	reflect "reflect"

	gomock "github.com/golang/mock/gomock"
)

// Producer is a mock of Producer interface.
type Producer struct {
	ctrl     *gomock.Controller
	recorder *ProducerMockRecorder
}

// ProducerMockRecorder is the mock recorder for Producer.
type ProducerMockRecorder struct {
	mock *Producer
}

// NewProducer creates a new mock instance.
func NewProducer(ctrl *gomock.Controller) *Producer {
	mock := &Producer{ctrl: ctrl}
	mock.recorder = &ProducerMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *Producer) EXPECT() *ProducerMockRecorder {
	return m.recorder
}

// Close mocks base method.
func (m *Producer) Close() error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Close")
	ret0, _ := ret[0].(error)
	return ret0
}

// Close indicates an expected call of Close.
func (mr *ProducerMockRecorder) Close() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Close", reflect.TypeOf((*Producer)(nil).Close))
}

// Send mocks base method.
func (m *Producer) Send(arg0, arg1 []byte) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Send", arg0, arg1)
	ret0, _ := ret[0].(error)
	return ret0
}

// Send indicates an expected call of Send.
func (mr *ProducerMockRecorder) Send(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Send", reflect.TypeOf((*Producer)(nil).Send), arg0, arg1)
}
