package mocks

import (
	"reflect"

	"github.com/golang/mock/gomock"

	"github.com/venafi/digicert-ca-connector/internal/app/domain"
)

// MockOptionsServices is a mock of ConnectorServices interface.
type MockOptionsServices struct {
	ctrl     *gomock.Controller
	recorder *MockOptionsServicesMockRecorder
}

// MockOptionsServicesMockRecorder is the mock recorder for MockOptionsServices.
type MockOptionsServicesMockRecorder struct {
	mock *MockOptionsServices
}

// NewMockOptionsServices creates a new mock instance.
func NewMockOptionsServices(ctrl *gomock.Controller) *MockOptionsServices {
	mock := &MockOptionsServices{ctrl: ctrl}
	mock.recorder = &MockOptionsServicesMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockOptionsServices) EXPECT() *MockOptionsServicesMockRecorder {
	return m.recorder
}

// GetOptions mocks base method.
func (m *MockOptionsServices) GetOptions(connection domain.Connection) ([]domain.ProductOption, []domain.ImportOption, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetOptions", connection)
	ret0 := ret[0].([]domain.ProductOption)
	ret1 := ret[1].([]domain.ImportOption)
	ret2, _ := ret[2].(error)
	return ret0, ret1, ret2
}

// GetOptions indicates an expected call of GetOptions.
func (mr *MockOptionsServicesMockRecorder) GetOptions(connection domain.Connection) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetOptions", reflect.TypeOf((*MockOptionsServices)(nil).GetOptions), connection)
}

// ValidateProduct mocks base method.
func (m *MockOptionsServices) ValidateProduct(connection domain.Connection, name string, product domain.Product) ([]domain.ProductError, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ValidateProduct", connection, name, product)
	ret0 := ret[0].([]domain.ProductError)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ValidateProduct indicates an expected call of GetOptions.
func (mr *MockOptionsServicesMockRecorder) ValidateProduct(connection domain.Connection, name string, product domain.Product) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ValidateProduct", reflect.TypeOf((*MockOptionsServices)(nil).ValidateProduct), connection, name, product)
}
