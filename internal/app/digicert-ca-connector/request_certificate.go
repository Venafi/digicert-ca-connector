package digicert_ca_connector

import (
	"fmt"
	"net/http"

	"github.com/venafi/digicert-ca-connector/internal/app/domain"

	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

// RequestCertificateRequest contains request details for submitting certificate request to Certificate Authority
type RequestCertificateRequest struct {
	Connection        domain.Connection      `json:"connection"`
	ValiditySeconds   int                    `json:"validitySeconds"`
	ProductOptionName string                 `json:"productOptionName"`
	Product           domain.Product         `json:"product"`
	Pkcs10Request     string                 `json:"pkcs10Request"`
	ProductDetails    *domain.ProductDetails `json:"productDetails"`
}

// RequestCertificateResponse contains certificate or/and order details for the submitted certificate request
type RequestCertificateResponse struct {
	CertificateDetails *domain.CertificateDetails `json:"certificateDetails"`
	OrderDetails       *domain.OrderDetails       `json:"orderDetails"`
}

// HandleRequestCertificate will submit certificate request to Certificate Authority
func (svc *WebhookService) HandleRequestCertificate(c echo.Context) error {
	req := RequestCertificateRequest{}
	if err := c.Bind(&req); err != nil {
		zap.L().Error("invalid request, failed to unmarshal json", zap.Error(err))
		return c.String(http.StatusBadRequest, fmt.Sprintf("failed to unmarshal json: %s", err.Error()))
	}

	cert, order, err := svc.Certificate.RequestCertificate(req.Connection, req.Pkcs10Request, req.Product, req.ProductOptionName, req.ValiditySeconds, req.ProductDetails)
	if err != nil {
		return c.String(http.StatusBadRequest, err.Error())
	}

	return c.JSON(http.StatusOK, &RequestCertificateResponse{
		CertificateDetails: cert,
		OrderDetails:       order,
	})
}
