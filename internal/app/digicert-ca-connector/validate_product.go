package digicert_ca_connector

import (
	"fmt"
	"net/http"

	"github.com/venafi/digicert-ca-connector/internal/app/domain"

	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

// ValidateProductRequest contains request details for validating specific product against Certificate Authority
type ValidateProductRequest struct {
	Connection  domain.Connection `json:"connection"`
	ProductName string            `json:"name"`
	Product     domain.Product    `json:"product"`
}

// ValidateProductResponse contains error details about invalid product attributes
type ValidateProductResponse struct {
	Errors []domain.ProductError `json:"errors"`
}

// HandleValidateProduct will validate specific product attributes against Certificate Authority
func (svc *WebhookService) HandleValidateProduct(c echo.Context) error {
	req := ValidateProductRequest{}
	if err := c.Bind(&req); err != nil {
		zap.L().Error("invalid request, failed to unmarshal json", zap.Error(err))
		return c.String(http.StatusBadRequest, fmt.Sprintf("failed to unmarshal json: %s", err.Error()))
	}

	productErrors, err := svc.Options.ValidateProduct(req.Connection, req.ProductName, req.Product)

	if err != nil {
		return c.String(http.StatusBadRequest, err.Error())
	}

	return c.JSON(http.StatusOK, &ValidateProductResponse{
		Errors: productErrors,
	})
}
