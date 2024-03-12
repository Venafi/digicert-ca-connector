package digicert_ca_connector

import (
	"fmt"
	"net/http"

	"github.com/venafi/digicert-ca-connector/internal/app/domain"

	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

// CheckCertificateRequest contains request details for retrieving certificate details for submitted certificate request from Certificate Authority
type CheckCertificateRequest struct {
	Connection domain.Connection `json:"connection"`
	ID         string            `json:"id"`
}

// HandleCheckCertificate will retrieve certificate details for submitted certificate request from Certificate Authority
func (svc *WebhookService) HandleCheckCertificate(c echo.Context) error {
	req := CheckCertificateRequest{}
	if err := c.Bind(&req); err != nil {
		zap.L().Error("invalid request, failed to unmarshal json", zap.Error(err))
		return c.String(http.StatusBadRequest, fmt.Sprintf("failed to unmarshal json: %s", err.Error()))
	}

	cert, err := svc.Certificate.CheckCertificate(req.Connection, req.ID)
	if err != nil {
		return c.String(http.StatusBadRequest, err.Error())
	}

	return c.JSON(http.StatusOK, cert)
}
