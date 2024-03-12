package mocks

import (
	"reflect"

	"github.com/golang/mock/gomock"

	"github.com/venafi/digicert-ca-connector/internal/app/domain"
)

// MockConnectorServices is a mock of ConnectorServices interface.
type MockConnectorServices struct {
	ctrl     *gomock.Controller
	recorder *MockConnectorServicesMockRecorder
}

// MockConnectorServicesMockRecorder is the mock recorder for MockConnectorServices.
type MockConnectorServicesMockRecorder struct {
	mock *MockConnectorServices
}

// NewMockConnectorServices creates a new mock instance.
func NewMockConnectorServices(ctrl *gomock.Controller) *MockConnectorServices {
	mock := &MockConnectorServices{ctrl: ctrl}
	mock.recorder = &MockConnectorServicesMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockConnectorServices) EXPECT() *MockConnectorServicesMockRecorder {
	return m.recorder
}

// TestConnection mocks base method.
func (m *MockConnectorServices) TestConnection(connection domain.Connection) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "TestConnection", connection)
	ret0, _ := ret[0].(error)
	return ret0
}

// TestConnection indicates an expected call of TestConnection.
func (mr *MockConnectorServicesMockRecorder) TestConnection(connection domain.Connection) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "TestConnection", reflect.TypeOf((*MockConnectorServices)(nil).TestConnection), connection)
}
