package digicert_ca_connector

import (
	"fmt"
	"net/http"

	"github.com/venafi/digicert-ca-connector/internal/app/domain"

	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

// TestConnectionRequest contains request details for testing connectivity to Certificate Authority
type TestConnectionRequest struct {
	Connection domain.Connection `json:"connection"`
}

// TestConnectionStatus status for testing connectivity
type TestConnectionStatus string

const (
	TestConnectionSuccess TestConnectionStatus = "SUCCESS"
	TestConnectionFailed  TestConnectionStatus = "FAILED"
)

// TestConnectionResponse contains test connectivity result
type TestConnectionResponse struct {
	Result  TestConnectionStatus `json:"result"`
	Message string               `json:"message"`
}

// HandleTestConnection will test connectivity to Certificate Authority
func (svc *WebhookService) HandleTestConnection(c echo.Context) error {
	req := TestConnectionRequest{}
	if err := c.Bind(&req); err != nil {
		zap.L().Error("invalid request, failed to unmarshal json", zap.Error(err))
		return c.String(http.StatusBadRequest, fmt.Sprintf("failed to unmarshal json: %s", err.Error()))
	}

	res := TestConnectionResponse{
		Result: TestConnectionFailed,
	}

	err := svc.Connections.TestConnection(req.Connection)
	if err != nil {
		zap.L().Error("error connecting to DigiCert Certificate Authority", zap.String("error", err.Error()))
		res.Message = fmt.Sprintf("failed to connect to DigiCert Certificate Authority: %s", err.Error())
	} else {
		res.Result = TestConnectionSuccess
		zap.L().Info("success connecting to DigiCert Certificate Authority")
	}
	return c.JSON(http.StatusOK, res)
}
