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
	getOptionsPath = "/v1/getOptions"
)

func (resp *GetOptionsResponse) unmarshal(body io.ReadCloser) error {
	data, err := io.ReadAll(body)
	if err != nil {
		return err
	}

	return json.Unmarshal(data, resp)
}

// TestHandleGetOptions ...
func TestHandleGetOptions(t *testing.T) {
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
		testGetOptions(t, whService, mockOptionsServices, e)
	})

	t.Run("invalid request no body", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		req := httptest.NewRequest(http.MethodPost, getOptionsPath, http.NoBody)
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		recorder := httptest.NewRecorder()

		err := whService.HandleGetOptions(e.NewContext(req, recorder))
		require.NoError(t, err)
		requireStatusBadRequest(t, "failed to unmarshal json", recorder.Result())
	})
}

func testGetOptions(t *testing.T, whService *WebhookService, mockOptionsServices *mocks.MockOptionsServices, e *echo.Echo) {
	recorder, ctx := setupPost(e, getOptionsPath, fmt.Sprintf(`{
			"connection": {
				"configuration": {
				    "serverUrl": "%s"
		       },
		       "credentials": {
		           "apiKey": "%s"
		       }
		   }
		}`, serverURL, apiKey))

	connection := buildConnection()

	mockOptionsServices.EXPECT().GetOptions(connection).DoAndReturn(func(connection domain.Connection) ([]domain.ProductOption, []domain.ImportOption, error) {
		return []domain.ProductOption{
				{
					Name:  "SSL Certificates",
					Types: []domain.ProductType{domain.ProductTypeSsl},
					Details: domain.ProductDetails{
						NameID:               "SSL Certificates ID",
						Hashes:               []string{"sha256", "sha512"},
						DefaultHashAlgorithm: "sha256",
						Organizations:        []int{1, 2},
					},
				},
				{
					Name:  "CodeSign Certificates",
					Types: []domain.ProductType{domain.ProductTypeCodeSign},
					Details: domain.ProductDetails{
						NameID:               "CodeSign Certificates ID",
						Hashes:               []string{"sha256", "sha512"},
						DefaultHashAlgorithm: "sha256",
						Organizations:        []int{1, 2},
					},
				},
			},
			[]domain.ImportOption{
				{
					Name:        "SSL Certificates",
					Description: "SSL Certificates available for import",
					Settings: domain.ImportSettings{
						NameID: "SSL Certificates ID",
					},
				},
			}, nil
	})

	err := whService.HandleGetOptions(ctx)
	require.NoError(t, err)

	response := recorder.Result()
	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(response.Body)

	require.NotNil(t, response)

	require.Equal(t, http.StatusOK, response.StatusCode)

	cr := &GetOptionsResponse{}
	err = cr.unmarshal(response.Body)
	require.NoError(t, err)
	require.Equal(t, len(cr.ProductOptions), 2)
	require.Equal(t, cr.ProductOptions[0].Name, "SSL Certificates")
	require.Equal(t, cr.ProductOptions[0].Types, []domain.ProductType{domain.ProductTypeSsl})
	require.Equal(t, cr.ProductOptions[0].Details.NameID, "SSL Certificates ID")
	require.Equal(t, cr.ProductOptions[0].Details.Hashes, []string{"sha256", "sha512"})
	require.Equal(t, cr.ProductOptions[0].Details.DefaultHashAlgorithm, "sha256")
	require.Equal(t, cr.ProductOptions[0].Details.Organizations, []int{1, 2})
	require.Equal(t, cr.ProductOptions[1].Name, "CodeSign Certificates")
	require.Equal(t, cr.ProductOptions[1].Types, []domain.ProductType{domain.ProductTypeCodeSign})
	require.Equal(t, cr.ProductOptions[1].Details.NameID, "CodeSign Certificates ID")
	require.Equal(t, cr.ProductOptions[1].Details.Hashes, []string{"sha256", "sha512"})
	require.Equal(t, cr.ProductOptions[1].Details.DefaultHashAlgorithm, "sha256")
	require.Equal(t, cr.ProductOptions[1].Details.Organizations, []int{1, 2})
	require.Equal(t, len(cr.ImportOptions), 1)
	require.Equal(t, cr.ImportOptions[0].Name, "SSL Certificates")
	require.Equal(t, cr.ImportOptions[0].Description, "SSL Certificates available for import")
	require.Equal(t, cr.ImportOptions[0].Settings.NameID, "SSL Certificates ID")
}
