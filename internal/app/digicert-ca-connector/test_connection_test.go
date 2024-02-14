package digicert_ca_connector

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/require"

	"github.com/venafi/digicert-ca-connector/internal/app/digicert-ca-connector/mocks"
	"github.com/venafi/digicert-ca-connector/internal/app/domain"
)

const (
	testConnectionPath = "/v1/testConnection"
	serverURL          = "https://digicert-test"
	apiKey             = "apiKey"
)

func (tcr *TestConnectionResponse) unmarshal(body io.ReadCloser) error {
	data, err := io.ReadAll(body)
	if err != nil {
		return err
	}

	return json.Unmarshal(data, tcr)
}

func setupPost(e *echo.Echo, path, body string) (*httptest.ResponseRecorder, echo.Context) {
	req := httptest.NewRequest(http.MethodPost, path, strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	recorder := httptest.NewRecorder()
	return recorder, e.NewContext(req, recorder)
}

func buildConnection() domain.Connection {
	return domain.Connection{
		Configuration: domain.Configuration{
			ServerURL: serverURL,
		},
		Credentials: domain.Credentials{
			ApiKey: apiKey,
		},
	}
}

func requireStatusBadRequest(t *testing.T, expectedErrorMessageInBody string, response *http.Response) {
	var err error

	require.NotNil(t, response)
	defer func(body io.ReadCloser) {
		err = body.Close()
		require.NoError(t, err)
	}(response.Body)

	require.Equal(t, http.StatusBadRequest, response.StatusCode)

	var data []byte
	data, err = io.ReadAll(response.Body)
	require.NoError(t, err)
	require.Contains(t, string(data), expectedErrorMessageInBody)
}

// TestHandleTestConnection ...
func TestHandleTestConnection(t *testing.T) {
	e := echo.New()
	require.NotNil(t, e)

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockConnectorServices := mocks.NewMockConnectorServices(ctrl)
	require.NotNil(t, mockConnectorServices)

	whService := NewWebhookService(mockConnectorServices, nil, nil)
	require.NotNil(t, whService)

	t.Parallel()

	t.Run("success", func(t *testing.T) {
		testTestConnection(t, whService, mockConnectorServices, e, true)
	})

	t.Run("failure", func(t *testing.T) {
		testTestConnection(t, whService, mockConnectorServices, e, false)
	})

	t.Run("invalid request no body", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		req := httptest.NewRequest(http.MethodPost, testConnectionPath, http.NoBody)
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		recorder := httptest.NewRecorder()

		err := whService.HandleTestConnection(e.NewContext(req, recorder))
		require.NoError(t, err)
		requireStatusBadRequest(t, "failed to unmarshal json", recorder.Result())
	})
}

func testTestConnection(t *testing.T, whService *WebhookService, mockConnectorServices *mocks.MockConnectorServices, e *echo.Echo, success bool) {
	recorder, ctx := setupPost(e, testConnectionPath, fmt.Sprintf(`{
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
	mockConnectorServices.EXPECT().TestConnection(connection).DoAndReturn(func(connection domain.Connection) error {
		if success {
			return nil
		}
		return fmt.Errorf("fail to connect")
	})

	err := whService.HandleTestConnection(ctx)
	require.NoError(t, err)

	response := recorder.Result()
	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(response.Body)

	require.NotNil(t, response)

	require.Equal(t, http.StatusOK, response.StatusCode)

	cr := &TestConnectionResponse{}
	err = cr.unmarshal(response.Body)
	require.NoError(t, err)
	if success {
		require.Equal(t, TestConnectionSuccess, cr.Result)
	} else {
		require.Equal(t, TestConnectionFailed, cr.Result)
		require.Equal(t, "failed to connect to DigiCert Certificate Authority: fail to connect", cr.Message)
	}
}
