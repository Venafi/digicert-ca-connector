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

// TestGetOptions ...
func TestGetOptions(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		testGetOptions(t)
	})
}

func testGetOptions(t *testing.T) {
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

	httpmock.RegisterResponder("GET", serverURL+getOrganizationsUri,
		func(req *http.Request) (*http.Response, error) {
			return httpmock.NewJsonResponse(http.StatusOK, &getOrganizationsResponse{
				Organizations: []organization{
					{
						ID:       1,
						Name:     "Org 1",
						Status:   "active",
						IsActive: true,
					},
					{
						ID:       2,
						Name:     "Org 2",
						Status:   "inactive",
						IsActive: false,
					},
				},
			})
		},
	)

	httpmock.RegisterResponder("GET", serverURL+getProductUri,
		func(req *http.Request) (*http.Response, error) {
			return httpmock.NewJsonResponse(http.StatusOK, &getProductDetails{
				ProductDetails: []digiCertProductDetails{
					{
						Name:            "SSL Certificates",
						NameID:          "SSL Certificates ID",
						CertificateType: "ssl_certificate",
						Hashes: signatureHashTypes{
							AllowedHashTypes: []hashType{
								{
									ID:   "sha256",
									Name: "SHA-256",
								},
								{
									ID:   "sha512",
									Name: "SHA-512",
								},
							},
							DefaultHashType: "sha256",
						},
					},
					{
						Name:            "CodeSign Certificates",
						NameID:          "CodeSign Certificates ID",
						CertificateType: "code_signing_certificate",
						Hashes: signatureHashTypes{
							AllowedHashTypes: []hashType{
								{
									ID:   "sha256",
									Name: "SHA-256",
								},
								{
									ID:   "sha512",
									Name: "SHA-512",
								},
							},
							DefaultHashType: "sha256",
						},
					},
				},
			})
		},
	)

	productOptions, _, err := NewOptionsService().GetOptions(connection)
	require.NoError(t, err)
	require.Len(t, productOptions, 2)
	require.Equal(t, productOptions[0].Name, "SSL Certificates")
	require.Equal(t, productOptions[0].Types, []domain.ProductType{domain.ProductTypeSsl})
	require.Equal(t, productOptions[0].Details.NameID, "SSL Certificates ID")
	require.Equal(t, productOptions[0].Details.Hashes, []string{"sha256", "sha512"})
	require.Equal(t, productOptions[0].Details.DefaultHashAlgorithm, "sha256")
	require.Equal(t, productOptions[0].Details.Organizations, []int{1})
	require.Equal(t, productOptions[1].Name, "CodeSign Certificates")
	require.Equal(t, productOptions[1].Types, []domain.ProductType{domain.ProductTypeCodeSign})
	require.Equal(t, productOptions[1].Details.NameID, "CodeSign Certificates ID")
	require.Equal(t, productOptions[1].Details.Hashes, []string{"sha256", "sha512"})
	require.Equal(t, productOptions[1].Details.DefaultHashAlgorithm, "sha256")
	require.Equal(t, productOptions[1].Details.Organizations, []int{1})
}
