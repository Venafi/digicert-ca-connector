package digicert_ca_connector

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/require"

	"github.com/venafi/digicert-ca-connector/internal/app/digicert-ca-connector/mocks"
	"github.com/venafi/digicert-ca-connector/internal/app/domain"
)

const (
	importCertificatesPath     = "/v1/importCertificates"
	lastProcessedCertificateID = "0"
	batchSize                  = 100
)

// TestHandleCheckCertificate ...
func TestHandleImportCertificates(t *testing.T) {
	e := echo.New()
	require.NotNil(t, e)

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockCertificateService := mocks.NewMockCertificateService(ctrl)
	require.NotNil(t, mockCertificateService)

	whService := NewWebhookService(nil, nil, mockCertificateService)
	require.NotNil(t, whService)

	t.Parallel()

	t.Run("completed", func(t *testing.T) {
		testRetrieveCertificates(t, whService, mockCertificateService, e, true)
	})

	t.Run("uncompleted", func(t *testing.T) {
		testRetrieveCertificates(t, whService, mockCertificateService, e, false)
	})

	t.Run("invalid request no body", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		req := httptest.NewRequest(http.MethodPost, importCertificatesPath, http.NoBody)
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		recorder := httptest.NewRecorder()

		err := whService.HandleImportCertificates(e.NewContext(req, recorder))
		require.NoError(t, err)
		requireStatusBadRequest(t, "failed to unmarshal json", recorder.Result())
	})
}

func testRetrieveCertificates(t *testing.T, whService *WebhookService, mockCertificateService *mocks.MockCertificateService, e *echo.Echo, complete bool) {
	recorder, ctx := setupPost(e, importCertificatesPath, fmt.Sprintf(`{
			"connection": {
				"configuration": {
				    "serverUrl": "%s"
		       },
		       "credentials": {
		           "apiKey": "%s"
		       }
		   },
           "option": {
				"name": "SSL Certificates",
				"description": "SSL Certificates are available for import",
                "settings": {
                    "nameId": "SSL Certificates ID"
                }
			},
			"configuration": {
				"includeExpiredCertificates": true
			},
			"lastProcessedCertificateId": "%s",
			"batchSize": %d
		}`, serverURL, apiKey, lastProcessedCertificateID, batchSize))

	connection := buildConnection()
	option := domain.ImportOption{
		Name:        "SSL Certificates",
		Description: "SSL Certificates are available for import",
		Settings: domain.ImportSettings{
			NameID: "SSL Certificates ID",
		},
	}
	importConfiguration := domain.ImportConfiguration{
		IncludeExpiredCertificates: true,
	}
	expectedImportDetails := &domain.ImportDetails{}
	mockCertificateService.EXPECT().RetrieveCertificates(connection, option, importConfiguration, lastProcessedCertificateID, batchSize).DoAndReturn(func(connection domain.Connection, option domain.ImportOption, configuration domain.ImportConfiguration, lastProcessedCertificateId string, batchSize int) (*domain.ImportDetails, error) {
		if complete {
			expectedImportDetails.ImportStatus = domain.ImportStatusCompleted
		} else {
			expectedImportDetails.ImportStatus = domain.ImportStatusUncompleted
		}
		expectedImportDetails.LastProcessedCertificateID = lastProcessedCertificateId
		expectedImportDetails.ImportCertificates = []domain.ImportCertificate{
			{
				ID:          "identifier1",
				Certificate: "Cert1",
				Chain:       []string{"Chain"},
			},
			{
				ID:          "identifier2",
				Certificate: "Cert2",
				Chain:       []string{"Chain"},
			}}
		return expectedImportDetails, nil
	})

	err := whService.HandleImportCertificates(ctx)
	require.NoError(t, err)

	response := recorder.Result()
	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(response.Body)

	require.NotNil(t, response)

	require.Equal(t, http.StatusOK, response.StatusCode)

	cd := &domain.ImportDetails{}
	data, err := io.ReadAll(response.Body)
	require.NoError(t, err)
	err = json.Unmarshal(data, cd)
	require.NoError(t, err)

	if complete {
		require.Equal(t, expectedImportDetails.ImportStatus, domain.ImportStatusCompleted)
	} else {
		require.Equal(t, expectedImportDetails.ImportStatus, domain.ImportStatusUncompleted)
	}
	require.Equal(t, expectedImportDetails.LastProcessedCertificateID, cd.LastProcessedCertificateID)
	require.Equal(t, reflect.DeepEqual(expectedImportDetails.ImportCertificates, cd.ImportCertificates), true)
}
