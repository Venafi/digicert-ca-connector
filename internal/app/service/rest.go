package service

import (
	"fmt"
	"net/http"

	"github.com/venafi/digicert-ca-connector/internal/app/domain"

	"github.com/go-resty/resty/v2"
)

// NewRestClient is a function that creates a resty client, to allow mocking and intercepting of HTTP requests
var NewRestClient = resty.New

func executeRequest(connection domain.Connection, requestBody any, uriPath string) (*resty.Response, error) {
	request := NewRestClient().R().SetHeader("Content-Type", "application/json").SetHeader("X-DC-DEVKEY", connection.Credentials.ApiKey)
	var resp *resty.Response
	var err error
	if requestBody != nil {
		resp, err = request.SetBody(requestBody).Post(connection.Configuration.ServerURL + uriPath)
	} else {
		resp, err = request.Get(connection.Configuration.ServerURL + uriPath)
	}

	if err != nil {
		return nil, err
	}

	if resp.StatusCode() != http.StatusOK && resp.StatusCode() != http.StatusCreated && resp.StatusCode() != http.StatusAccepted {
		return nil, fmt.Errorf(string(resp.Body()))
	}
	return resp, nil
}
