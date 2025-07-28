package digicert_ca_connector

import (
	"fmt"
	"github.com/labstack/echo/v4"
	"github.com/venafi/digicert-ca-connector/internal/app/domain"
	"go.uber.org/zap"
	"net/http"
)

// RevokeCertificateRequest contains request details for submitting certificate revocation request to Certificate Authority
type RevokeCertificateRequest struct {
	Connection                domain.Connection                `json:"connection"`
	CertificateRevocationData domain.CertificateRevocationData `json:"certificateRevocationData"`
	Reason                    int                              `json:"reason"`
}

type RevokeCertificateResponse struct {
	RevocationStatus domain.RevocationStatus `json:"revocationStatus"`
	ErrorMessage     *string                 `json:"errorMessage"`
}

// HandleRevokeCertificate will submit certificate revocation request to Certificate Authority
func (svc *WebhookService) HandleRevokeCertificate(c echo.Context) error {
	req := RevokeCertificateRequest{}
	if err := c.Bind(&req); err != nil {
		zap.L().Error("invalid request, failed to unmarshal json", zap.Error(err))
		return c.String(http.StatusBadRequest, fmt.Sprintf("failed to unmarshal json: %s", err.Error()))
	}

	resp, err := svc.Certificate.RevokeCertificate(req.Connection, req.CertificateRevocationData.SerialNumber, req.Reason)
	if err != nil {
		return c.String(http.StatusBadRequest, err.Error())
	}

	res := RevokeCertificateResponse{RevocationStatus: resp.Status, ErrorMessage: resp.ErrorMessage}

	return c.JSON(http.StatusOK, &res)
}
