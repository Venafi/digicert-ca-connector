// Package web contains the web server and registered routes
package web

import (
	"bytes"
	"context"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"io"
	"io/ioutil"
	"net/http"
	"os"

	"github.com/go-jose/go-jose/v4"
	"github.com/labstack/echo/v4"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

// WebhookService interfaces for the connector operation functions
type WebhookService interface {
	HandleTestConnection(c echo.Context) error
	HandleGetOptions(c echo.Context) error
	HandleValidateProduct(c echo.Context) error
	HandleRequestCertificate(c echo.Context) error
	HandleCheckOrder(c echo.Context) error
	HandleCheckCertificate(c echo.Context) error
	HandleImportCertificates(c echo.Context) error
	HandleRevokeCertificate(c echo.Context) error
}

// ConfigureHTTPServers creates an HTTP server with standard middleware and a system HTTP server with health and metrics endpoints
// returns the echo engine for serving API
func ConfigureHTTPServers(lifecycle fx.Lifecycle, shutdowner fx.Shutdowner) (*echo.Echo, error) {
	e := echo.New()

	lifecycle.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			go func() {
				if err := e.Start(":8080"); err != nil && err != http.ErrServerClosed {
					zap.L().Error("failed to start echo server", zap.Error(err))
					if err = shutdowner.Shutdown(); err != nil {
						zap.L().Error("fx shutdown error", zap.Error(err))
					}
				}
			}()
			return nil
		},
		OnStop: func(ctx context.Context) error {
			return e.Shutdown(ctx)
		},
	})

	return e, nil
}

// RegisterHandlers adds the method handlers for the supported routes
func RegisterHandlers(e *echo.Echo, whService WebhookService) error {
	e.GET("/healthz", func(c echo.Context) error {
		return c.String(http.StatusOK, "OK")
	})

	g := e.Group("/v1")
	addPayloadEncryptionMiddleware(g)
	g.POST("/testconnection", whService.HandleTestConnection)
	g.POST("/getoptions", whService.HandleGetOptions)
	g.POST("/validateproduct", whService.HandleValidateProduct)
	g.POST("/requestcertificate", whService.HandleRequestCertificate)
	g.POST("/checkorder", whService.HandleCheckOrder)
	g.POST("/checkcertificate", whService.HandleCheckCertificate)
	g.POST("/importcertificates", whService.HandleImportCertificates)
	g.POST("/revokecertificate", whService.HandleRevokeCertificate)

	return nil
}

func addPayloadEncryptionMiddleware(g *echo.Group) {
	privateKeyPemData, err := os.ReadFile("/keys/payload-encryption-key.pem")
	if err != nil {
		zap.L().Fatal("payload encryption key not found or readable", zap.Error(err))
	}
	p, _ := pem.Decode(privateKeyPemData)
	if p == nil {
		zap.L().Fatal("payload encryption key not in PEM format")
	}
	pk, err := x509.ParsePKCS1PrivateKey(p.Bytes)
	if err != nil {
		zap.L().Fatal("payload encryption key not properly encoded", zap.Error(err))
	}
	zap.L().Info("adding payload encryption middleware")
	g.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			req := c.Request()
			body, err := ioutil.ReadAll(req.Body)
			if err != nil {
				return echo.NewHTTPError(http.StatusUnauthorized, "failed to read request body")
			}
			object, err := jose.ParseEncrypted(
				string(body),
				[]jose.KeyAlgorithm{
					jose.ED25519, jose.RSA1_5, jose.RSA_OAEP, jose.RSA_OAEP_256, jose.A128KW, jose.A192KW, jose.A256KW,
					jose.DIRECT, jose.ECDH_ES, jose.ECDH_ES_A128KW, jose.ECDH_ES_A192KW, jose.ECDH_ES_A256KW, jose.A128GCMKW,
					jose.A192GCMKW, jose.A256GCMKW, jose.PBES2_HS256_A128KW, jose.PBES2_HS384_A192KW, jose.PBES2_HS512_A256KW,
				},
				[]jose.ContentEncryption{
					jose.A128CBC_HS256, jose.A192CBC_HS384, jose.A256CBC_HS512, jose.A128GCM, jose.A192GCM, jose.A256GCM,
				},
			)
			if err != nil {
				return echo.NewHTTPError(http.StatusUnauthorized, "invalid encrypted payload")
			}
			decrypted, err := object.Decrypt(pk)
			if err != nil {
				return echo.NewHTTPError(http.StatusUnauthorized, "decryption failed")
			}
			// CWE-400: Limit decompressed JWE payload to 16 MiB to prevent decompression bombs
			const maxDecryptedSize = 16 * 1024 * 1024
			if len(decrypted) > maxDecryptedSize {
				return errors.New("decrypted payload exceeds maximum allowed size")
			}
			req.Body = io.NopCloser(bytes.NewReader(decrypted))
			return next(c)
		}
	})
}
