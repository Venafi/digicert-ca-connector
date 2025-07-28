package mocks

import (
	"reflect"

	"github.com/golang/mock/gomock"

	"github.com/venafi/digicert-ca-connector/internal/app/domain"
)

// MockCertificateService is a mock of CertificateService interface.
type MockCertificateService struct {
	ctrl     *gomock.Controller
	recorder *MockCertificateServiceMockRecorder
}

// MockCertificateServiceMockRecorder is the mock recorder for MockCertificateService.
type MockCertificateServiceMockRecorder struct {
	mock *MockCertificateService
}

// NewMockCertificateService creates a new mock instance.
func NewMockCertificateService(ctrl *gomock.Controller) *MockCertificateService {
	mock := &MockCertificateService{ctrl: ctrl}
	mock.recorder = &MockCertificateServiceMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockCertificateService) EXPECT() *MockCertificateServiceMockRecorder {
	return m.recorder
}

// RequestCertificate mocks base method.
func (m *MockCertificateService) RequestCertificate(connection domain.Connection, pkcs10Request string, product domain.Product, productOptionName string, validitySeconds int, productDetails *domain.ProductDetails) (*domain.CertificateDetails, *domain.OrderDetails, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "RequestCertificate", connection, pkcs10Request, product, productOptionName, validitySeconds, productDetails)
	ret0, _ := ret[0].(*domain.CertificateDetails)
	ret1, _ := ret[1].(*domain.OrderDetails)
	ret2, _ := ret[2].(error)
	return ret0, ret1, ret2
}

// RequestCertificate indicates an expected call of MockCertificateService.
func (mr *MockCertificateServiceMockRecorder) RequestCertificate(connection domain.Connection, pkcs10Request string, product domain.Product, productOptionName string, validitySeconds int, productDetails *domain.ProductDetails) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "RequestCertificate", reflect.TypeOf((*MockCertificateService)(nil).RequestCertificate), connection, pkcs10Request, product, productOptionName, validitySeconds, productDetails)
}

// CheckOrder mocks base method.
func (m *MockCertificateService) CheckOrder(connection domain.Connection, id string) (*domain.OrderDetails, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "CheckOrder", connection, id)
	ret0, _ := ret[0].(*domain.OrderDetails)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// CheckOrder indicates an expected call of MockCertificateService.
func (mr *MockCertificateServiceMockRecorder) CheckOrder(connection domain.Connection, id string) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CheckOrder", reflect.TypeOf((*MockCertificateService)(nil).CheckOrder), connection, id)
}

// CheckCertificate mocks base method.
func (m *MockCertificateService) CheckCertificate(connection domain.Connection, id string) (*domain.CertificateDetails, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "CheckCertificate", connection, id)
	ret0, _ := ret[0].(*domain.CertificateDetails)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// CheckCertificate indicates an expected call of MockCertificateService.
func (mr *MockCertificateServiceMockRecorder) CheckCertificate(connection domain.Connection, id string) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CheckCertificate", reflect.TypeOf((*MockCertificateService)(nil).CheckCertificate), connection, id)
}

// RetrieveCertificates mocks base method.
func (m *MockCertificateService) RetrieveCertificates(connection domain.Connection, option domain.ImportOption, configuration domain.ImportConfiguration, startCursor string, batchSize int) (*domain.ImportDetails, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "RetrieveCertificates", connection, option, configuration, startCursor, batchSize)
	ret0, _ := ret[0].(*domain.ImportDetails)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// RetrieveCertificates indicates an expected call of MockCertificateService.
func (mr *MockCertificateServiceMockRecorder) RetrieveCertificates(connection domain.Connection, option domain.ImportOption, configuration domain.ImportConfiguration, startCursor string, batchSize int) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "RetrieveCertificates", reflect.TypeOf((*MockCertificateService)(nil).RetrieveCertificates), connection, option, configuration, startCursor, batchSize)
}

// RevokeCertificate mocks base method.
func (m *MockCertificateService) RevokeCertificate(connection domain.Connection, serialNumber string, reasonCode int) (*domain.RevocationDetails, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "RevokeCertificate", connection, serialNumber, reasonCode)
	ret0, _ := ret[0].(*domain.RevocationDetails)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// RevokeCertificate indicates an expected call of MockCertificateService.
func (mr *MockCertificateServiceMockRecorder) RevokeCertificate(connection domain.Connection, serialNumber string, reasonCode int) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "RevokeCertificate", reflect.TypeOf((*MockCertificateService)(nil).RevokeCertificate), connection, serialNumber, reasonCode)
}
