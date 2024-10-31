// Code generated by MockGen. DO NOT EDIT.
// Source: proto_test.go
//
// Generated by this command:
//
//	mockgen -source=proto_test.go -destination=proto_mocks_test.go -package=proto
//

// Package proto is a generated GoMock package.
package proto

import (
	reflect "reflect"

	gomock "go.uber.org/mock/gomock"
)

// Mockwriter is a mock of writer interface.
type Mockwriter struct {
	isgomock struct{}
	ctrl     *gomock.Controller
	recorder *MockwriterMockRecorder
}

// MockwriterMockRecorder is the mock recorder for Mockwriter.
type MockwriterMockRecorder struct {
	mock *Mockwriter
}

// NewMockwriter creates a new mock instance.
func NewMockwriter(ctrl *gomock.Controller) *Mockwriter {
	mock := &Mockwriter{ctrl: ctrl}
	mock.recorder = &MockwriterMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *Mockwriter) EXPECT() *MockwriterMockRecorder {
	return m.recorder
}

// Write mocks base method.
func (m *Mockwriter) Write(p []byte) (int, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Write", p)
	ret0, _ := ret[0].(int)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Write indicates an expected call of Write.
func (mr *MockwriterMockRecorder) Write(p any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Write", reflect.TypeOf((*Mockwriter)(nil).Write), p)
}

// Mockreader is a mock of reader interface.
type Mockreader struct {
	isgomock struct{}
	ctrl     *gomock.Controller
	recorder *MockreaderMockRecorder
}

// MockreaderMockRecorder is the mock recorder for Mockreader.
type MockreaderMockRecorder struct {
	mock *Mockreader
}

// NewMockreader creates a new mock instance.
func NewMockreader(ctrl *gomock.Controller) *Mockreader {
	mock := &Mockreader{ctrl: ctrl}
	mock.recorder = &MockreaderMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *Mockreader) EXPECT() *MockreaderMockRecorder {
	return m.recorder
}

// Read mocks base method.
func (m *Mockreader) Read(p []byte) (int, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Read", p)
	ret0, _ := ret[0].(int)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Read indicates an expected call of Read.
func (mr *MockreaderMockRecorder) Read(p any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Read", reflect.TypeOf((*Mockreader)(nil).Read), p)
}