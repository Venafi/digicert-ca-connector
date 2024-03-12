// Package main implements the application main function.
package main

import (
	"github.com/venafi/digicert-ca-connector/cmd/digicert-ca-connector/app"
)

func main() {
	app.New().Run()
}
