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
	validateProductPath = "/v1/validateproduct"
)

func (tcr *ValidateProductResponse) unmarshal(body io.ReadCloser) error {
	data, err := io.ReadAll(body)
	if err != nil {
		return err
	}

	return json.Unmarshal(data, tcr)
}

// TestHandleValidateProduct ...
func TestHandleValidateProduct(t *testing.T) {
	e := echo.New()
	require.NotNil(t, e)

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockOptionsServices := mocks.NewMockOptionsServices(ctrl)
	require.NotNil(t, mockOptionsServices)

	whService := NewWebhookService(nil, mockOptionsServices, nil)
	require.NotNil(t, whService)

	t.Parallel()

	t.Run("success", func(t *testing.T) {
		testValidateProduct(t, whService, mockOptionsServices, e, true)
	})

	t.Run("failure", func(t *testing.T) {
		testValidateProduct(t, whService, mockOptionsServices, e, false)
	})

	t.Run("invalid request no body", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		req := httptest.NewRequest(http.MethodPost, validateProductPath, http.NoBody)
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		recorder := httptest.NewRecorder()

		err := whService.HandleValidateProduct(e.NewContext(req, recorder))
		require.NoError(t, err)
		requireStatusBadRequest(t, "failed to unmarshal json", recorder.Result())
	})
}

func testValidateProduct(t *testing.T, whService *WebhookService, mockOptionsServices *mocks.MockOptionsServices, e *echo.Echo, success bool) {
	recorder, ctx := setupPost(e, validateProductPath, fmt.Sprintf(`{
			"connection": {
				"configuration": {
				    "serverUrl": "%s"
		       },
		       "credentials": {
		           "apiKey": "%s"
		       }
		   },
           "name": "%s",
           "product": {
               "nameId": "%s",
               "hashAlgorithm": "%s",
               "organizationId": %d
           }
		}`, serverURL, apiKey, productOptionName, productNameId, productHashAlgorithm, productOrganizationId))

	connection := buildConnection()
	product := domain.Product{
		NameID:         productNameId,
		HashAlgorithm:  productHashAlgorithm,
		OrganizationID: productOrganizationId,
	}
	mockOptionsServices.EXPECT().ValidateProduct(connection, productOptionName, product).DoAndReturn(func(connection domain.Connection, productOptionName string, product domain.Product) ([]domain.ProductError, error) {
		if success {
			return nil, nil
		}
		return []domain.ProductError{
			{
				AttributeName:  "nameId",
				AttributeValue: productNameId,
			},
			{
				AttributeName:  "hashAlgorithm",
				AttributeValue: productHashAlgorithm,
			},
		}, nil
	})

	err := whService.HandleValidateProduct(ctx)
	require.NoError(t, err)

	response := recorder.Result()
	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(response.Body)

	require.NotNil(t, response)

	require.Equal(t, http.StatusOK, response.StatusCode)

	cr := &ValidateProductResponse{}
	err = cr.unmarshal(response.Body)
	require.NoError(t, err)
	if success {
		require.Nil(t, cr.Errors)
	} else {
		require.Equal(t, len(cr.Errors), 2)
		require.Equal(t, cr.Errors[0].AttributeName, "nameId")
		require.Equal(t, cr.Errors[0].AttributeValue, productNameId)
		require.Equal(t, cr.Errors[1].AttributeName, "hashAlgorithm")
		require.Equal(t, cr.Errors[1].AttributeValue, productHashAlgorithm)
	}
}
