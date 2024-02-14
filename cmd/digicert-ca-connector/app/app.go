// Package app represents the application logic.
package app

import (
	connector "github.com/venafi/digicert-ca-connector/internal/app/digicert-ca-connector"
	"github.com/venafi/digicert-ca-connector/internal/app/service"
	"github.com/venafi/digicert-ca-connector/internal/handler/web"

	"go.uber.org/fx"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// New ...
func New() *fx.App {
	var logger *zap.Logger

	app := fx.New(
		fx.Provide(
			configureLogger,
			web.ConfigureHTTPServers,
			fx.Annotate(service.NewConnectionService, fx.As(new(connector.ConnectionService))),
			fx.Annotate(service.NewOptionsService, fx.As(new(connector.OptionsService))),
			fx.Annotate(service.NewCertificateService, fx.As(new(connector.CertificateService))),
			fx.Annotate(connector.NewWebhookService, fx.As(new(web.WebhookService))),
		),
		fx.Invoke(
			web.RegisterHandlers,
		),
		fx.Populate(&logger),
	)

	logger.Info("CA connector starting")

	return app
}

func configureLogger() (*zap.Logger, error) {
	loggerConfig := zap.NewProductionConfig()
	loggerConfig.Level = zap.NewAtomicLevelAt(zap.DebugLevel)
	loggerConfig.EncoderConfig = zap.NewProductionEncoderConfig()
	loggerConfig.EncoderConfig.TimeKey = "time"
	loggerConfig.EncoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder
	loggerConfig.EncoderConfig.EncodeTime = zapcore.RFC3339TimeEncoder
	logger, err := loggerConfig.Build()
	if err != nil {
		return nil, err
	}

	if err != nil {
		return nil, err
	}
	zap.ReplaceGlobals(logger)
	zap.RedirectStdLog(zap.L())
	return zap.L(), nil
}
