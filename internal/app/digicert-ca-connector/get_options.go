package digicert_ca_connector

import (
	"fmt"
	"net/http"

	"github.com/venafi/digicert-ca-connector/internal/app/domain"

	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

// GetOptionsRequest contains request details for retrieving product and import options from Certificate Authority
type GetOptionsRequest struct {
	Connection domain.Connection `json:"connection"`
}

// GetOptionsResponse contains product and import options retrieved from Certificate Authority
type GetOptionsResponse struct {
	ProductOptions []domain.ProductOption `json:"productOptions"`
	ImportOptions  []domain.ImportOption  `json:"importOptions"`
}

// HandleGetOptions will retrieve product and import options from Certificate Authority
func (svc *WebhookService) HandleGetOptions(c echo.Context) error {
	req := GetOptionsRequest{}
	if err := c.Bind(&req); err != nil {
		zap.L().Error("invalid request, failed to unmarshal json", zap.Error(err))
		return c.String(http.StatusBadRequest, fmt.Sprintf("failed to unmarshal json: %s", err.Error()))
	}

	po, io, err := svc.Options.GetOptions(req.Connection)
	if err != nil {
		return c.String(http.StatusBadRequest, err.Error())
	}

	return c.JSON(http.StatusOK, &GetOptionsResponse{
		ProductOptions: po,
		ImportOptions:  io,
	})
}
