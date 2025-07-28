package digicert_ca_connector

import (
	"encoding/json"
	"fmt"
	"github.com/golang/mock/gomock"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/require"
	"github.com/venafi/digicert-ca-connector/internal/app/digicert-ca-connector/mocks"
	"github.com/venafi/digicert-ca-connector/internal/app/domain"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

const (
	revokeCertificatePath   = "/v1/revokecertificate"
	serialNumber            = "test-serial"
	fingerprint             = "test-fingerprint"
	caCertificateIdentifier = "test-cert-identifier"
	caOrderIdentifier       = "test-order-identifier"
	issuerDN                = "test-issuer"
	certificateContent      = "--------- BEGIN CERTIFICATE --------\nTest Cert\n--------- END CERTIFICATE --------"
	reason                  = 1
)

func (resp *RevokeCertificateResponse) unmarshal(body io.ReadCloser) error {
	data, err := io.ReadAll(body)
	if err != nil {
		return err
	}

	return json.Unmarshal(data, resp)
}

func TestHandleRevokeCertificate(t *testing.T) {
	e := echo.New()
	require.NotNil(t, e)

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockCertificateService := mocks.NewMockCertificateService(ctrl)
	require.NotNil(t, mockCertificateService)

	whService := NewWebhookService(nil, nil, mockCertificateService)
	require.NotNil(t, whService)

	t.Parallel()

	t.Run("success revocation", func(t *testing.T) {
		testRevokeCertificate(t, whService, mockCertificateService, e, true)
	})

	t.Run("failed revocation", func(t *testing.T) {
		testRevokeCertificate(t, whService, mockCertificateService, e, false)
	})

	t.Run("invalid request no body", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		req := httptest.NewRequest(http.MethodPost, revokeCertificatePath, http.NoBody)
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		recorder := httptest.NewRecorder()

		err := whService.HandleRevokeCertificate(e.NewContext(req, recorder))
		require.NoError(t, err)
		requireStatusBadRequest(t, "failed to unmarshal json", recorder.Result())
	})
}

func testRevokeCertificate(t *testing.T, whService *WebhookService, mockCertificateService *mocks.MockCertificateService, e *echo.Echo, success bool) {
	crd := domain.CertificateRevocationData{
		SerialNumber:            serialNumber,
		Fingerprint:             fingerprint,
		CaCertificateIdentifier: caCertificateIdentifier,
		CaOrderIdentifier:       caOrderIdentifier,
		IssuerDN:                issuerDN,
		CertificateContent:      certificateContent,
	}
	crdJson, _ := json.Marshal(crd)
	recorder, ctx := setupPost(e, revokeCertificatePath, fmt.Sprintf(`{
			"connection": {
				"configuration": {
				    "serverUrl": "%s"
		       },
		       "credentials": {
		           "apiKey": "%s"
		       }
		   },
           "certificateRevocationData": %s,
           "reason": %d
		}`, serverURL, apiKey, crdJson, reason))

	connection := buildConnection()
	var expectedRevocationDetails domain.RevocationDetails
	mockCertificateService.EXPECT().RevokeCertificate(connection, serialNumber, reason).DoAndReturn(func(connection domain.Connection, serialNumber string, reason int) (*domain.RevocationDetails, error) {
		if success {
			expectedRevocationDetails.Status = domain.RevocationStatusSubmitted
		} else {
			expectedRevocationDetails.Status = domain.RevocationStatusFailed
			errMsg := "failed to submit certificate revocation request"
			expectedRevocationDetails.ErrorMessage = &errMsg
		}
		return &expectedRevocationDetails, nil
	})

	err := whService.HandleRevokeCertificate(ctx)
	require.NoError(t, err)

	response := recorder.Result()
	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(response.Body)

	require.NotNil(t, response)

	require.Equal(t, http.StatusOK, response.StatusCode)

	cr := &RevokeCertificateResponse{}
	err = cr.unmarshal(response.Body)
	require.NoError(t, err)

	if success {
		require.Equal(t, expectedRevocationDetails.Status, cr.RevocationStatus)
		require.Nil(t, cr.ErrorMessage)
	} else {
		require.Equal(t, expectedRevocationDetails.Status, cr.RevocationStatus)
		require.Equal(t, *expectedRevocationDetails.ErrorMessage, *cr.ErrorMessage)
	}
}
