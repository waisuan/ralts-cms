// Code generated by MockGen. DO NOT EDIT.
// Source: repository.go

// Package machines is a generated GoMock package.
package machines

import (
	reflect "reflect"

	gomock "github.com/golang/mock/gomock"
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

// Create mocks base method.
func (m_2 *MockRepository) Create(m *Machine) (*Machine, error) {
	m_2.ctrl.T.Helper()
	ret := m_2.ctrl.Call(m_2, "Create", m)
	ret0, _ := ret[0].(*Machine)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Create indicates an expected call of Create.
func (mr *MockRepositoryMockRecorder) Create(m interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Create", reflect.TypeOf((*MockRepository)(nil).Create), m)
}

// DeleteBySerialNumber mocks base method.
func (m *MockRepository) DeleteBySerialNumber(serialNumber string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "DeleteBySerialNumber", serialNumber)
	ret0, _ := ret[0].(error)
	return ret0
}

// DeleteBySerialNumber indicates an expected call of DeleteBySerialNumber.
func (mr *MockRepositoryMockRecorder) DeleteBySerialNumber(serialNumber interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DeleteBySerialNumber", reflect.TypeOf((*MockRepository)(nil).DeleteBySerialNumber), serialNumber)
}

// GetBySerialNumber mocks base method.
func (m *MockRepository) GetBySerialNumber(serialNumber string) (*Machine, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetBySerialNumber", serialNumber)
	ret0, _ := ret[0].(*Machine)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetBySerialNumber indicates an expected call of GetBySerialNumber.
func (mr *MockRepositoryMockRecorder) GetBySerialNumber(serialNumber interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetBySerialNumber", reflect.TypeOf((*MockRepository)(nil).GetBySerialNumber), serialNumber)
}

// Query mocks base method.
func (m *MockRepository) Query(limit, offset int, sortField string, reversedOrder bool) ([]Machine, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Query", limit, offset, sortField, reversedOrder)
	ret0, _ := ret[0].([]Machine)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Query indicates an expected call of Query.
func (mr *MockRepositoryMockRecorder) Query(limit, offset, sortField, reversedOrder interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Query", reflect.TypeOf((*MockRepository)(nil).Query), limit, offset, sortField, reversedOrder)
}

// Update mocks base method.
func (m_2 *MockRepository) Update(m *Machine) (*Machine, error) {
	m_2.ctrl.T.Helper()
	ret := m_2.ctrl.Call(m_2, "Update", m)
	ret0, _ := ret[0].(*Machine)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Update indicates an expected call of Update.
func (mr *MockRepositoryMockRecorder) Update(m interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Update", reflect.TypeOf((*MockRepository)(nil).Update), m)
}
