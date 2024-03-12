package service

import (
	"net/http"
	"testing"

	"github.com/go-resty/resty/v2"
	"github.com/golang/mock/gomock"
	"github.com/jarcoal/httpmock"
	"github.com/stretchr/testify/require"

	"github.com/venafi/digicert-ca-connector/internal/app/domain"
)

const (
	serverURL = "https://digicert-test"
	apiKey    = "apiKey"
)

// TestConnect ...
func TestConnect(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		testConnect(t, http.StatusOK)
	})

	t.Run("failure", func(t *testing.T) {
		testConnect(t, http.StatusBadRequest)
	})
}

func testConnect(t *testing.T, httpStatus int) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	connection := buildConnection()

	// override the resty constructor to intercept HTTPS traffic
	savedRestCtor := NewRestClient
	defer func() { NewRestClient = savedRestCtor }()
	NewRestClient = func() *resty.Client {
		client := resty.New()
		httpmock.ActivateNonDefault(client.GetClient())
		return client
	}
	defer httpmock.DeactivateAndReset()

	body := ""
	if httpStatus != http.StatusOK {
		body = "failure"
	}

	httpmock.RegisterResponder("GET", serverURL+testConnectionUri,
		func(req *http.Request) (*http.Response, error) {
			return httpmock.NewJsonResponse(httpStatus, body)
		},
	)
	connector := NewConnectionService()

	err := connector.TestConnection(connection)
	if httpStatus == http.StatusOK {
		require.NoError(t, err)
	} else {
		require.Error(t, err)
	}
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
