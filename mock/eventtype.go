// Code generated by MockGen. DO NOT EDIT.
// Source: github.com/serverless/event-gateway/event (interfaces: Service)

// Package mock is a generated GoMock package.
package mock

import (
	gomock "github.com/golang/mock/gomock"
	event "github.com/serverless/event-gateway/event"
	reflect "reflect"
)

// MockEventTypeService is a mock of Service interface
type MockEventTypeService struct {
	ctrl     *gomock.Controller
	recorder *MockEventTypeServiceMockRecorder
}

// MockEventTypeServiceMockRecorder is the mock recorder for MockEventTypeService
type MockEventTypeServiceMockRecorder struct {
	mock *MockEventTypeService
}

// NewMockEventTypeService creates a new mock instance
func NewMockEventTypeService(ctrl *gomock.Controller) *MockEventTypeService {
	mock := &MockEventTypeService{ctrl: ctrl}
	mock.recorder = &MockEventTypeServiceMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockEventTypeService) EXPECT() *MockEventTypeServiceMockRecorder {
	return m.recorder
}

// CreateEventType mocks base method
func (m *MockEventTypeService) CreateEventType(arg0 *event.Type) (*event.Type, error) {
	ret := m.ctrl.Call(m, "CreateEventType", arg0)
	ret0, _ := ret[0].(*event.Type)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// CreateEventType indicates an expected call of CreateEventType
func (mr *MockEventTypeServiceMockRecorder) CreateEventType(arg0 interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CreateEventType", reflect.TypeOf((*MockEventTypeService)(nil).CreateEventType), arg0)
}

// DeleteEventType mocks base method
func (m *MockEventTypeService) DeleteEventType(arg0 string, arg1 event.TypeName) error {
	ret := m.ctrl.Call(m, "DeleteEventType", arg0, arg1)
	ret0, _ := ret[0].(error)
	return ret0
}

// DeleteEventType indicates an expected call of DeleteEventType
func (mr *MockEventTypeServiceMockRecorder) DeleteEventType(arg0, arg1 interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DeleteEventType", reflect.TypeOf((*MockEventTypeService)(nil).DeleteEventType), arg0, arg1)
}

// GetEventType mocks base method
func (m *MockEventTypeService) GetEventType(arg0 string, arg1 event.TypeName) (*event.Type, error) {
	ret := m.ctrl.Call(m, "GetEventType", arg0, arg1)
	ret0, _ := ret[0].(*event.Type)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetEventType indicates an expected call of GetEventType
func (mr *MockEventTypeServiceMockRecorder) GetEventType(arg0, arg1 interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetEventType", reflect.TypeOf((*MockEventTypeService)(nil).GetEventType), arg0, arg1)
}

// GetEventTypes mocks base method
func (m *MockEventTypeService) GetEventTypes(arg0 string) (event.Types, error) {
	ret := m.ctrl.Call(m, "GetEventTypes", arg0)
	ret0, _ := ret[0].(event.Types)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetEventTypes indicates an expected call of GetEventTypes
func (mr *MockEventTypeServiceMockRecorder) GetEventTypes(arg0 interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetEventTypes", reflect.TypeOf((*MockEventTypeService)(nil).GetEventTypes), arg0)
}

// UpdateEventType mocks base method
func (m *MockEventTypeService) UpdateEventType(arg0 *event.Type) (*event.Type, error) {
	ret := m.ctrl.Call(m, "UpdateEventType", arg0)
	ret0, _ := ret[0].(*event.Type)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// UpdateEventType indicates an expected call of UpdateEventType
func (mr *MockEventTypeServiceMockRecorder) UpdateEventType(arg0 interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "UpdateEventType", reflect.TypeOf((*MockEventTypeService)(nil).UpdateEventType), arg0)
}
