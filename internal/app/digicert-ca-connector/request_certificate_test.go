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
	requestCertificatePath = "/v1/certificaterequest"
	productOptionName      = "SSL Certificates"
	productNameId          = "SSL Certificates ID"
	productHashAlgorithm   = "sha256"
	productOrganizationId  = 1
	pkcs10Request          = "CSR"
	validitySeconds        = 300
)

func (resp *RequestCertificateResponse) unmarshal(body io.ReadCloser) error {
	data, err := io.ReadAll(body)
	if err != nil {
		return err
	}

	return json.Unmarshal(data, resp)
}

// TestHandleTestConnection ...
func TestHandleRequestCertificate(t *testing.T) {
	e := echo.New()
	require.NotNil(t, e)

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockCertificateService := mocks.NewMockCertificateService(ctrl)
	require.NotNil(t, mockCertificateService)

	whService := NewWebhookService(nil, nil, mockCertificateService)
	require.NotNil(t, whService)

	t.Parallel()

	t.Run("success cert details", func(t *testing.T) {
		testRequestCertificate(t, whService, mockCertificateService, e, true, false)
	})

	t.Run("success order details", func(t *testing.T) {
		testRequestCertificate(t, whService, mockCertificateService, e, true, true)
	})

	t.Run("failure cert details", func(t *testing.T) {
		testRequestCertificate(t, whService, mockCertificateService, e, false, false)
	})

	t.Run("failure order details", func(t *testing.T) {
		testRequestCertificate(t, whService, mockCertificateService, e, false, true)
	})

	t.Run("invalid request no body", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		req := httptest.NewRequest(http.MethodPost, requestCertificatePath, http.NoBody)
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		recorder := httptest.NewRecorder()

		err := whService.HandleRequestCertificate(e.NewContext(req, recorder))
		require.NoError(t, err)
		requireStatusBadRequest(t, "failed to unmarshal json", recorder.Result())
	})
}

func testRequestCertificate(t *testing.T, whService *WebhookService, mockCertificateService *mocks.MockCertificateService, e *echo.Echo, success bool, orderDetails bool) {
	pd := domain.ProductDetails{
		NameID:               productNameId,
		Hashes:               []string{productHashAlgorithm},
		DefaultHashAlgorithm: productHashAlgorithm,
		Organizations:        []int{productOrganizationId},
	}
	pdJson, _ := json.Marshal(pd)
	recorder, ctx := setupPost(e, requestCertificatePath, fmt.Sprintf(`{
			"connection": {
				"configuration": {
				    "serverUrl": "%s"
		       },
		       "credentials": {
		           "apiKey": "%s"
		       }
		   },
           "productOptionName": "%s",
           "product": {
               "nameId": "%s",
               "hashAlgorithm": "%s",
               "organizationId": %d
           },
           "pkcs10Request": "%s",
           "validitySeconds": %d,
           "productDetails": %s
		}`, serverURL, apiKey, productOptionName, productNameId, productHashAlgorithm, productOrganizationId, pkcs10Request, validitySeconds, pdJson))

	connection := buildConnection()
	po := domain.Product{
		NameID:         productNameId,
		HashAlgorithm:  productHashAlgorithm,
		OrganizationID: productOrganizationId,
	}
	var expectedCertDetails domain.CertificateDetails
	var expectedOrderDetails domain.OrderDetails
	mockCertificateService.EXPECT().RequestCertificate(connection, pkcs10Request, po, productOptionName, validitySeconds, &pd).DoAndReturn(func(connection domain.Connection, pkcs10Request string, product domain.Product, productOptionName string, validitySeconds int, productDetails *domain.ProductDetails) (*domain.CertificateDetails, *domain.OrderDetails, error) {
		if success {
			if orderDetails {
				expectedOrderDetails.ID = "OrderID"
				expectedOrderDetails.CertificateID = "CertID"
				expectedOrderDetails.Status = domain.OrderStatusCompleted
			} else {
				expectedCertDetails.ID = "CertID"
				expectedCertDetails.Status = domain.CertificateStatusIssued
				expectedCertDetails.Certificate = "Cert"
				expectedCertDetails.Chain = []string{"Chain"}
			}
		} else {
			if orderDetails {
				expectedOrderDetails.Status = domain.OrderStatusFailed
				expectedOrderDetails.ErrorMessage = "fail to connect"
			} else {
				expectedCertDetails.Status = domain.CertificateStatusFailed
				expectedCertDetails.ErrorMessage = "fail to connect"
			}
		}
		return &expectedCertDetails, &expectedOrderDetails, nil
	})

	err := whService.HandleRequestCertificate(ctx)
	require.NoError(t, err)

	response := recorder.Result()
	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(response.Body)

	require.NotNil(t, response)

	require.Equal(t, http.StatusOK, response.StatusCode)

	cr := &RequestCertificateResponse{}
	err = cr.unmarshal(response.Body)
	require.NoError(t, err)

	if success {
		if orderDetails {
			require.Equal(t, expectedOrderDetails.ID, cr.OrderDetails.ID)
			require.Equal(t, expectedOrderDetails.CertificateID, cr.OrderDetails.CertificateID)
			require.Equal(t, expectedOrderDetails.Status, cr.OrderDetails.Status)
		} else {
			require.Equal(t, expectedCertDetails.ID, cr.CertificateDetails.ID)
			require.Equal(t, domain.CertificateStatusIssued, cr.CertificateDetails.Status)
			require.Equal(t, expectedCertDetails.Certificate, cr.CertificateDetails.Certificate)
			require.Equal(t, expectedCertDetails.Chain, cr.CertificateDetails.Chain)
			require.Empty(t, cr.CertificateDetails.ErrorMessage)
		}
	} else {
		if orderDetails {
			require.Equal(t, expectedOrderDetails.Status, cr.OrderDetails.Status)
			require.Equal(t, expectedOrderDetails.ErrorMessage, cr.OrderDetails.ErrorMessage)
		} else {
			require.Equal(t, domain.CertificateStatusFailed, cr.CertificateDetails.Status)
			require.Equal(t, expectedCertDetails.ErrorMessage, cr.CertificateDetails.ErrorMessage)
		}
	}
}
