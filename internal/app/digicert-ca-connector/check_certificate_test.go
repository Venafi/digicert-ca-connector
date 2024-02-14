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
	certificateId        = "certificateId"
	checkCertificatePath = "/v1/checkcertificate"
)

// TestHandleCheckCertificate ...
func TestHandleCheckCertificate(t *testing.T) {
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
		testCheckCertificate(t, whService, mockCertificateService, e, true)
	})

	t.Run("failure", func(t *testing.T) {
		testCheckCertificate(t, whService, mockCertificateService, e, false)
	})

	t.Run("invalid request no body", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		req := httptest.NewRequest(http.MethodPost, checkCertificatePath, http.NoBody)
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		recorder := httptest.NewRecorder()

		err := whService.HandleCheckCertificate(e.NewContext(req, recorder))
		require.NoError(t, err)
		requireStatusBadRequest(t, "failed to unmarshal json", recorder.Result())
	})
}

func testCheckCertificate(t *testing.T, whService *WebhookService, mockCertificateService *mocks.MockCertificateService, e *echo.Echo, success bool) {
	recorder, ctx := setupPost(e, checkCertificatePath, fmt.Sprintf(`{
			"connection": {
				"configuration": {
				    "serverUrl": "%s"
		       },
		       "credentials": {
		           "apiKey": "%s"
		       }
		   },
          "id": "%s"
		}`, serverURL, apiKey, certificateId))

	connection := buildConnection()
	expectedCertDetails := &domain.CertificateDetails{}
	mockCertificateService.EXPECT().CheckCertificate(connection, certificateId).DoAndReturn(func(connection domain.Connection, id string) (*domain.CertificateDetails, error) {
		if success {
			expectedCertDetails.ID = "CertID"
			expectedCertDetails.Status = domain.CertificateStatusIssued
			expectedCertDetails.Certificate = "Cert"
			expectedCertDetails.Chain = []string{"Chain"}
		} else {
			expectedCertDetails.Status = domain.CertificateStatusFailed
			expectedCertDetails.ErrorMessage = "fail to connect"
		}
		return expectedCertDetails, nil
	})

	err := whService.HandleCheckCertificate(ctx)
	require.NoError(t, err)

	response := recorder.Result()
	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(response.Body)

	require.NotNil(t, response)

	require.Equal(t, http.StatusOK, response.StatusCode)

	cd := &domain.CertificateDetails{}
	data, err := io.ReadAll(response.Body)
	require.NoError(t, err)
	err = json.Unmarshal(data, cd)
	require.NoError(t, err)

	if success {
		require.Equal(t, expectedCertDetails.ID, cd.ID)
		require.Equal(t, domain.CertificateStatusIssued, cd.Status)
		require.Equal(t, expectedCertDetails.Certificate, cd.Certificate)
		require.Equal(t, expectedCertDetails.Chain, cd.Chain)
		require.Empty(t, cd.ErrorMessage)
	} else {
		require.Equal(t, domain.CertificateStatusFailed, cd.Status)
		require.Equal(t, expectedCertDetails.ErrorMessage, cd.ErrorMessage)
	}
}
