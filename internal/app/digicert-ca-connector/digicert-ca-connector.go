package digicert_ca_connector

import (
	"github.com/venafi/digicert-ca-connector/internal/app/domain"
)

// ConnectionService ...
type ConnectionService interface {
	TestConnection(connection domain.Connection) error
}

// OptionsService ...
type OptionsService interface {
	GetOptions(connection domain.Connection) ([]domain.ProductOption, []domain.ImportOption, error)
	ValidateProduct(connection domain.Connection, name string, product domain.Product) ([]domain.ProductError, error)
}

// CertificateService ...
type CertificateService interface {
	RequestCertificate(connection domain.Connection, pkcs10Request string, product domain.Product, productOptionName string, validitySeconds int) (*domain.CertificateDetails, *domain.OrderDetails, error)
	CheckOrder(connection domain.Connection, id string) (*domain.OrderDetails, error)
	CheckCertificate(connection domain.Connection, id string) (*domain.CertificateDetails, error)
	RetrieveCertificates(connection domain.Connection, option domain.ImportOption, configuration domain.ImportConfiguration, startCursor string, batchSize int) (*domain.ImportDetails, error)
}

// WebhookService ...
type WebhookService struct {
	Connections ConnectionService
	Options     OptionsService
	Certificate CertificateService
}

// NewWebhookService will return a new service
func NewWebhookService(connections ConnectionService, options OptionsService, certificate CertificateService) *WebhookService {
	return &WebhookService{
		Connections: connections,
		Options:     options,
		Certificate: certificate,
	}
}
