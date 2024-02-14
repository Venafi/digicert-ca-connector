package service

import (
	"github.com/venafi/digicert-ca-connector/internal/app/domain"
)

const (
	testConnectionUri = "/organization"
)

// Connector ...
type Connector struct {
}

// NewConnectionService will return a new webhook service
func NewConnectionService() *Connector {
	return &Connector{}
}

// TestConnection will test connection against a Certificate Authority
func (cs *Connector) TestConnection(connection domain.Connection) error {
	_, err := executeRequest(connection, nil, testConnectionUri)
	return err
}
