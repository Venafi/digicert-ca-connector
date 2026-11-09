package digicert_ca_connector

import (
	"errors"

	"github.com/labstack/echo/v4"
)

var errEmptyRequestBody = errors.New("request body is empty")

// bindRequest binds the JSON request body into req. echo's default binder
// treats an empty body as success and leaves req at its zero value, which
// would send an empty connection to DigiCert; reject it instead.
func bindRequest(c echo.Context, req any) error {
	if c.Request().ContentLength == 0 {
		return errEmptyRequestBody
	}
	return c.Bind(req)
}
