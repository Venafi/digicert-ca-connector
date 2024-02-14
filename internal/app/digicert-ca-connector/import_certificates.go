package digicert_ca_connector

import (
	"fmt"
	"net/http"

	"github.com/venafi/digicert-ca-connector/internal/app/domain"

	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

// ImportCertificatesRequest contains request details for retrieving certificates available for import in TLSPC from Certificate Authority
type ImportCertificatesRequest struct {
	Connection                 domain.Connection          `json:"connection"`
	Option                     domain.ImportOption        `json:"option"`
	Configuration              domain.ImportConfiguration `json:"configuration"`
	LastProcessedCertificateID string                     `json:"lastProcessedCertificateId"`
	BatchSize                  int                        `json:"batchSize"`
}

// HandleImportCertificates will retrieve certificates available for import in TLSPC from Certificate Authority
func (svc *WebhookService) HandleImportCertificates(c echo.Context) error {
	req := ImportCertificatesRequest{}
	if err := c.Bind(&req); err != nil {
		zap.L().Error("invalid request, failed to unmarshal json", zap.Error(err))
		return c.String(http.StatusBadRequest, fmt.Sprintf("failed to unmarshal json: %s", err.Error()))
	}

	res, err := svc.Certificate.RetrieveCertificates(req.Connection, req.Option, req.Configuration, req.LastProcessedCertificateID, req.BatchSize)
	if err != nil {
		zap.L().Error("failed to retrieve certificates from Certificate Authority", zap.Error(err))
		return c.String(http.StatusBadRequest, fmt.Sprintf("failed to retrieve certificates from Certificate Authority: %s", err.Error()))
	}
	return c.JSON(http.StatusOK, res)
}
