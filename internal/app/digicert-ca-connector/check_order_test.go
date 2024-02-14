package digicert_ca_connector

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/require"
	"github.com/venafi/digicert-ca-connector/internal/app/digicert-ca-connector/mocks"

	"github.com/venafi/digicert-ca-connector/internal/app/domain"
)

const (
	orderID        = "orderId"
	checkOrderPath = "/v1/checkOrder"
)

// TestHandleCheckOrder ...
func TestHandleCheckOrder(t *testing.T) {
	e := echo.New()
	require.NotNil(t, e)

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockCertificateService := mocks.NewMockCertificateService(ctrl)
	require.NotNil(t, mockCertificateService)

	whService := NewWebhookService(nil, nil, mockCertificateService)
	require.NotNil(t, whService)

	t.Parallel()

	t.Run("success", func(t *testing.T) {
		testCheckOrder(t, whService, mockCertificateService, e, true)
	})

	t.Run("failure", func(t *testing.T) {
		testCheckOrder(t, whService, mockCertificateService, e, false)
	})

	t.Run("invalid request no body", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		req := httptest.NewRequest(http.MethodPost, checkOrderPath, http.NoBody)
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		recorder := httptest.NewRecorder()

		err := whService.HandleCheckOrder(e.NewContext(req, recorder))
		require.NoError(t, err)
		requireStatusBadRequest(t, "failed to unmarshal json", recorder.Result())
	})
}

func testCheckOrder(t *testing.T, whService *WebhookService, mockCertificateService *mocks.MockCertificateService, e *echo.Echo, success bool) {
	recorder, ctx := setupPost(e, checkOrderPath, fmt.Sprintf(`{
			"connection": {
				"configuration": {
				    "serverUrl": "%s"
		       },
		       "credentials": {
		           "apiKey": "%s"
		       }
		   },
          "id": "%s"
		}`, serverURL, apiKey, orderID))

	connection := buildConnection()
	expectedDetails := &domain.OrderDetails{}
	mockCertificateService.EXPECT().CheckOrder(connection, orderID).DoAndReturn(func(connection domain.Connection, id string) (*domain.OrderDetails, error) {
		if success {
			expectedDetails.ID = orderID
			expectedDetails.Status = domain.OrderStatusCompleted
			expectedDetails.CertificateID = "CertId"
		} else {
			expectedDetails.Status = domain.OrderStatusFailed
			expectedDetails.ErrorMessage = "fail to connect"
		}
		return expectedDetails, nil
	})

	err := whService.HandleCheckOrder(ctx)
	require.NoError(t, err)

	response := recorder.Result()
	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(response.Body)

	require.NotNil(t, response)

	require.Equal(t, http.StatusOK, response.StatusCode)

	cd := &domain.OrderDetails{}
	data, err := io.ReadAll(response.Body)
	require.NoError(t, err)
	err = json.Unmarshal(data, cd)
	require.NoError(t, err)

	if success {
		require.Equal(t, expectedDetails.ID, cd.ID)
		require.Equal(t, domain.OrderStatusCompleted, cd.Status)
		require.Equal(t, expectedDetails.CertificateID, cd.CertificateID)
		require.Empty(t, cd.ErrorMessage)
	} else {
		require.Equal(t, domain.OrderStatusFailed, cd.Status)
		require.Equal(t, expectedDetails.ErrorMessage, cd.ErrorMessage)
	}
}
