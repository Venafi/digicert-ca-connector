package service

import (
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"testing"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/golang/mock/gomock"
	"github.com/jarcoal/httpmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/venafi/digicert-ca-connector/internal/app/domain"
)

const (
	productHashAlgorithm  = "sha256"
	productOrganizationId = 1
	productOptionName     = "SSL Certificates"
	pkcs10Request         = "-----BEGIN CERTIFICATE REQUEST-----\nMIICpzCCAY8CAQAwHDEaMBgGA1UEAwwRZGlnaWNlcnQtdGVzdC5jb20wggEiMA0G\nCSqGSIb3DQEBAQUAA4IBDwAwggEKAoIBAQCG6uyC8F+FCpedCtHcFxUnB8lVVXM0\nIjw5rrkMCgJ3VyMLpWjerlk2ZYnMkvDeOPTHrXQsnqocsYJwvx6zYep6xpvxNTEj\nb1/+t39iPX43nUqvu19YCipq6hEvW21R2+VsiVJH8SVT40hhn7IT52vkUzRlpWlO\nuCl3tV9Z1+z8YguDz3waBLDqMLaYX+ikZ5srYMnxtN4L1r2XN5bzjAEwoSwlDLqe\nlNkY/jCIgT5UQ9UkzYvwyEgLvgNJFsGMGz1h34IsbIftfEyE/lhFSW+CagXHR/x3\neZqPVncet4zC2qyELBzTopBwn4iukaUGqYm+RylpyE26QUfzoPuj35xzAgMBAAGg\nRjBEBgkqhkiG9w0BCQ4xNzA1MDMGA1UdEQQsMCqCEWRpZ2ljZXJ0LXRlc3QuY29t\nghV3d3cuZGlnaWNlcnQtdGVzdC5jb20wDQYJKoZIhvcNAQELBQADggEBADdCQlpc\n4mxBS2ExkJ8CxqKsRBtj7RsayHCBrJpZgFMQ0fzQmjdoTWoJOJWvPcopfd6pw7Q1\n9iVFbyKsIvM/nLaMjn6cNYfM+QK48aFBAJm/UJ3XGECrfla74J8Mp8AtGRfbuBWq\nUfFnEpzPOfrSrPbrtPPIEVKvUSpulNnwWxLsyS8v3Y1OutkiULNhB1jxTua6JPyP\n33uCGCwBk8y1n/aEQ6UOD3rQlaZ/OmXV2GrqLY9fATjoOpjlRKy/txKabJ+jf2i/\nrLcJtICkBKXybquUOqseoMkptAqURXvaQhfNDuIMIV9utoT+qzFPmIrbiiN6eT9s\n+9Xxm0chARb3lr0=\n-----END CERTIFICATE REQUEST-----"
	ee_cert               = "-----BEGIN CERTIFICATE-----\nMIIEyDCCA7CgAwIBAgIQAnI/g3S06aSDmy49qtcoizANBgkqhkiG9w0BAQsFADBs\nMQswCQYDVQQGEwJVUzEVMBMGA1UEChMMRGlnaUNlcnQgSW5jMRkwFwYDVQQLExB3\nd3cuZGlnaWNlcnQuY29tMSswKQYDVQQDEyJEaWdpQ2VydCBUZXN0IEludGVybWVk\naWF0ZSBSb290IENBMB4XDTIzMTExMzE0MjE1OFoXDTIzMTEyNDE0MjE1N1owaTEL\nMAkGA1UEBhMCVVMxDTALBgNVBAgTBFV0YWgxFzAVBgNVBAcTDlNhbHQgTGFrZSBD\naXR5MRUwEwYDVQQKEwxWZW5hZmksIEluYy4xGzAZBgNVBAMTEnZlbmFmaS5leGFt\ncGxlLmNvbTCCASIwDQYJKoZIhvcNAQEBBQADggEPADCCAQoCggEBAMaA+8Ou18Ku\nm01aCZGoNBoNlj+m6Hwl0uVyPsanLGaazoz++iZMo5kPO9JpjJV9xNkbjNruFKFb\nQ6WtLkGLXkPvs/cPlcsO1yI0qK2urWvUZBeOkyhS4agg4XzhFD2djdlP0G2XpSyh\nJb0Ng5fFUmgXqOFLsgHszMk1/beoblx432sGTY3yOt58bIw80g/UsOIFniRQWtQz\nb1S18VZIGL77mkIlcEdesEOgFofkruGlb1TcHSKizbmBB34DQTl3ULyMHxw9MARf\nhr2BNe/pdUqOFpdM5fYffkyWjN/wMXgxxcOalczrVJjFYasyO3fDrI4Q5AY/YVpM\nlqxaOfmCo7ECAwEAAaOCAWcwggFjMB8GA1UdIwQYMBaAFPBhYIIrlI9DOZFs+a8q\n0tkei4HsMB0GA1UdDgQWBBSUNx31zLHgAMR+KSP0sWY6u9KPKTAdBgNVHREEFjAU\nghJ2ZW5hZmkuZXhhbXBsZS5jb20wQQYDVR0gBDowODA2BglghkgBhv1sAQEwKTAn\nBggrBgEFBQcCARYbaHR0cDovL3d3dy5kaWdpY2VydC5jb20vQ1BTMA4GA1UdDwEB\n/wQEAwIFoDAdBgNVHSUEFjAUBggrBgEFBQcDAQYIKwYBBQUHAwIwgYEGCCsGAQUF\nBwEBBHUwczAlBggrBgEFBQcwAYYZaHR0cDovL29jc3B4LmRpZ2ljZXJ0LmNvbTBK\nBggrBgEFBQcwAoY+aHR0cDovL2NhY2VydHMuZGlnaWNlcnQuY29tL0RpZ2lDZXJ0\nVGVzdEludGVybWVkaWF0ZVJvb3RDQS5jcnQwDAYDVR0TAQH/BAIwADANBgkqhkiG\n9w0BAQsFAAOCAQEAns6h7Ajc24Dr5/Ffc7Y3v9ULwUYGa/zNeiNNNSnd6rEpR3nv\nTllhFnFCjgUxNwXOcvSdSj6kROjj67TN+ZU5ubANF8UWtt1kJvdl88lza6NCk3Ap\nQxKeq40ncxvx9Sfh8+6ySe7GX6o47Iug9K3W/LocYVXi2MFAz04AmsM2xayi5ZlI\nfAtgTCEwwDfeEESV/pNaCarLu2/EvtyFzEn4ISsjq2oMmtq2u015dFFbwdofF358\nSF47SM0YQ/8lZ1tesdPLi3GJVk1+AYfg1DcIzDN6eDFtwZqmiKBdETvavz/gh1Vn\nmqs/1SUqQdjAryGdxzKlMkIDXVnrwe4JuSyD2w==\n-----END CERTIFICATE-----"
	intermediate_cert     = "-----BEGIN CERTIFICATE-----\nMIIGNTCCBR2gAwIBAgIEEaqqqjANBgkqhkiG9w0BAQUFADBfMQswCQYDVQQGEwJV\nUzEVMBMGA1UEChMMRGlnaUNlcnQgSW5jMRkwFwYDVQQLExB3d3cuZGlnaWNlcnQu\nY29tMR4wHAYDVQQDExVEaWdpQ2VydCBUZXN0IFJvb3QgQ0EwHhcNMDYxMTEwMDAw\nMDAwWhcNMzExMTEwMDAwMDAwWjBsMQswCQYDVQQGEwJVUzEVMBMGA1UEChMMRGln\naUNlcnQgSW5jMRkwFwYDVQQLExB3d3cuZGlnaWNlcnQuY29tMSswKQYDVQQDEyJE\naWdpQ2VydCBUZXN0IEludGVybWVkaWF0ZSBSb290IENBMIIBIjANBgkqhkiG9w0B\nAQEFAAOCAQ8AMIIBCgKCAQEAtOPuUF002sMiYiWjnII5YbI+FoGcbSRwKrN7PXVj\nwhXSxM8RiJqANHsrc5wPatKn/sNxIQdK8sUhXDgFrRFHrvb1LIOW+w7zuWtmMYgK\ny5cLPPOMm1SbwEzbOGDqa7yjFArkip5qSnQGpn+xQF5FFvX0xZaIedKuVKqn2KVm\nTdyyJjsMc2u2rLrazVvkBTuwpUvZScyoQ+UIh3+67HK5KPgh7uBwxKaQKzM6JVhZ\nFh30lBtzpYRn37uyDNh22hjJ3lNpi7NCUb3vY39e2/XYuaINRx6lpw2V+b1SQHRM\n2/o5cuxB+bNprd2MatZ1juRuyu71L7xjo/yDNMQitU4PnQIDAQABo4IC6jCCAuYw\nDgYDVR0PAQH/BAQDAgGGMIIBxgYDVR0gBIIBvTCCAbkwggG1BgtghkgBhv1sAQMA\nAjCCAaQwOgYIKwYBBQUHAgEWLmh0dHA6Ly93d3cuZGlnaWNlcnQuY29tL3NzbC1j\ncHMtcmVwb3NpdG9yeS5odG0wggFkBggrBgEFBQcCAjCCAVYeggFSAEEAbgB5ACAA\ndQBzAGUAIABvAGYAIAB0AGgAaQBzACAAQwBlAHIAdABpAGYAaQBjAGEAdABlACAA\nYwBvAG4AcwB0AGkAdAB1AHQAZQBzACAAYQBjAGMAZQBwAHQAYQBuAGMAZQAgAG8A\nZgAgAHQAaABlACAARABpAGcAaQBDAGUAcgB0ACAAQwBQAC8AQwBQAFMAIABhAG4A\nZAAgAHQAaABlACAAUgBlAGwAeQBpAG4AZwAgAFAAYQByAHQAeQAgAEEAZwByAGUA\nZQBtAGUAbgB0ACAAdwBoAGkAYwBoACAAbABpAG0AaQB0ACAAbABpAGEAYgBpAGwA\naQB0AHkAIABhAG4AZAAgAGEAcgBlACAAaQBuAGMAbwByAHAAbwByAGEAdABlAGQA\nIABoAGUAcgBlAGkAbgAgAGIAeQAgAHIAZQBmAGUAcgBlAG4AYwBlAC4wDwYDVR0T\nAQH/BAUwAwEB/zA4BggrBgEFBQcBAQQsMCowKAYIKwYBBQUHMAGGHGh0dHA6Ly9v\nY3NwdGVzdC5kaWdpY2VydC5jb20wfwYDVR0fBHgwdjA5oDegNYYzaHR0cDovL2Ny\nbDN0ZXN0LmRpZ2ljZXJ0LmNvbS9EaWdpQ2VydFRlc3RSb290Q0EuY3JsMDmgN6A1\nhjNodHRwOi8vY3JsNHRlc3QuZGlnaWNlcnQuY29tL0RpZ2lDZXJ0VGVzdFJvb3RD\nQS5jcmwwHQYDVR0OBBYEFPBhYIIrlI9DOZFs+a8q0tkei4HsMB8GA1UdIwQYMBaA\nFEawcgj8NeX6+v+d3lEQbmKVXdewMA0GCSqGSIb3DQEBBQUAA4IBAQCmPJ4TonbJ\n6g//M5ZLcuSc86RVARxVWUrB9oYZJRY+UjqiOusgqdhabTmuwf9Rjac5xUt6luUj\nIO0RRmskkEjCtUXjg/mtgM5BaBmTgnJvTQjyhb5+SPtVE4JlurASuLaa5CoTf0Mb\nrft+WPCk1BJMQmQiv2ptz00x9i5oysMyf5g3h8qMqV1kp1sWw4wsBd3XmcF6GrPJ\nvRmrdKZXmBnnDEe/nxSkqlzJXB41aIeiw1HhOQL2kxzdAAW58HI248Yi6MzfGIGY\nXGA7S3AodOTNDNmIxI9OSa6Rh8vsL8u5LxTknQbkzGTU+p97Ul1U1uWXVx6dW13Z\nwtADm8F+PrC8\n-----END CERTIFICATE-----"
	root_cert             = "-----BEGIN CERTIFICATE-----\nMIIDnDCCAoSgAwIBAgIBETANBgkqhkiG9w0BAQUFADBfMQswCQYDVQQGEwJVUzEV\nMBMGA1UEChMMRGlnaUNlcnQgSW5jMRkwFwYDVQQLExB3d3cuZGlnaWNlcnQuY29t\nMR4wHAYDVQQDExVEaWdpQ2VydCBUZXN0IFJvb3QgQ0EwHhcNMDYxMTEwMDAwMDAw\nWhcNMzExMTEwMDAwMDAwWjBfMQswCQYDVQQGEwJVUzEVMBMGA1UEChMMRGlnaUNl\ncnQgSW5jMRkwFwYDVQQLExB3d3cuZGlnaWNlcnQuY29tMR4wHAYDVQQDExVEaWdp\nQ2VydCBUZXN0IFJvb3QgQ0EwggEiMA0GCSqGSIb3DQEBAQUAA4IBDwAwggEKAoIB\nAQC2mEEv1NiVWb1x7GlRVwmW7tAUQhslTr+5Iz8tHrUq1l1+7rAxjkLzovibesr+\norXuL++zlpBAQKxIOQ1T9Kw8m+OKDtAjRiMhP4Mx6O2Qpe4N3Pras2pCPGToXrKf\n/68lPmt52Fnqd8ISoaBh0+i+SUWM2aNm6e+JFq6IQ/iE2crOXBHaRpv4/IOMCfrT\n6zAaFnsmWoUjGc6ISqb2nwsYMMOCZtH57ygc54GcIp7t6mmJ3S/Myewtkkk+AGrm\nhRAgi8/eE6eU++jQoGrZ8UfgYZahTSZkJHZtRj+m9sSUsMX2Lw4Uxk2gUUkdNHvo\nOdzd+sBLmiw5z6vI9d0YYfwBAgMBAAGjYzBhMA4GA1UdDwEB/wQEAwIBhjAPBgNV\nHRMBAf8EBTADAQH/MB0GA1UdDgQWBBRGsHII/DXl+vr/nd5REG5ilV3XsDAfBgNV\nHSMEGDAWgBRGsHII/DXl+vr/nd5REG5ilV3XsDANBgkqhkiG9w0BAQUFAAOCAQEA\nWcypG3UOkkFw+FEtQmXQDxPBWmS36KwQ64myJXnqcd41ZskYjyCE62iXd2qfQOQ0\naoTkbcIo3Ov7RX9M5+m3kpzZmlHHwef0ePd5p1dtVsmnR22TXdmpyxPDOLtYz7wd\n3DTG2G5fUN2/dgeTK8mITonetrVOkpVx8WtJkMGgVN5Dhy6gVYw0XpNfweyPNacq\nu0BwrelLn5qTBXCYwg7IWFP2Ca34Xr2tLcQ17zE+PX51TonA7RdB4eOZ2JE6cJp9\n5D0dyY/RjQvQpn8d7ZjSaHq0HzBMwcXkVMcoKjhOpmwoJz/sJzlt7WFpjd+xyNEr\nChW/tdOxL+vy0HBs7NYzkQ==\n-----END CERTIFICATE-----"
	certBatchSize         = 2
)

func TestRequestCertificate(t *testing.T) {
	t.Run("successRequestCertificate", func(t *testing.T) {
		testCertificateRequest(t, http.StatusOK, false)
	})

	t.Run("successRequestCertificateOrderDetails", func(t *testing.T) {
		testCertificateRequest(t, http.StatusOK, true)
	})

	t.Run("failureRequestCertificate", func(t *testing.T) {
		testCertificateRequest(t, http.StatusBadRequest, false)
	})

	t.Run("successCheckCertificate", func(t *testing.T) {
		testCheckCertificateData(t, http.StatusOK)
	})

	t.Run("failureCheckCertificate", func(t *testing.T) {
		testCheckCertificateData(t, http.StatusBadRequest)
	})

	t.Run("successCheckOrder", func(t *testing.T) {
		testCheckOrderData(t, http.StatusOK)
	})

	t.Run("failureCheckOrder", func(t *testing.T) {
		testCheckOrderData(t, http.StatusBadRequest)
	})

	t.Run("completeRetrieveCertificates", func(t *testing.T) {
		testRetrieveCertificateData(t, http.StatusOK, certBatchSize, true, true)
	})

	t.Run("uncompletedRetrieveCertificates", func(t *testing.T) {
		testRetrieveCertificateData(t, http.StatusOK, certBatchSize, false, true)
	})

	t.Run("uncompletedRetrieveCertificatesNoExpired", func(t *testing.T) {
		testRetrieveCertificateData(t, http.StatusOK, certBatchSize, false, false)
	})

	t.Run("errorRetrieveCertificates", func(t *testing.T) {
		testRetrieveCertificateData(t, http.StatusBadRequest, certBatchSize, true, false)
	})
}

func testCertificateRequest(t *testing.T, httpStatus int, orderDetails bool) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	connection := buildConnection()

	certRequest := newCertificateRequestBody{
		Certificate: certificate{
			CommonName: "digicert-test.com",
			DnsNames:   []string{"digicert-test.com"},
			Csr:        pkcs10Request,
			ServerPlatform: serverPlatform{
				ID: -1,
			},
			SignatureHash: productHashAlgorithm,
		},
		Organization: digicertOrganization{
			ID: productOrganizationId,
		},
		CustomExpirationDate: time.Now().Add(time.Second * time.Duration(300)).Format(digicertDateFormat),
	}

	// override the resty constructor to intercept HTTPS traffic
	savedRestCtor := NewRestClient
	defer func() { NewRestClient = savedRestCtor }()
	NewRestClient = func() *resty.Client {
		client := resty.New()
		httpmock.ActivateNonDefault(client.GetClient())
		return client
	}
	defer httpmock.DeactivateAndReset()

	httpmock.RegisterResponder("POST", serverURL+fmt.Sprintf(orderCertificateUri, "ssl_private_id"),
		func(req *http.Request) (*http.Response, error) {
			data, err := io.ReadAll(req.Body)
			assert.NoError(t, err)

			reqBody := &newCertificateRequestBody{}
			err = json.Unmarshal(data, reqBody)
			assert.NoError(t, err)
			assert.Equal(t, reqBody.Certificate.CommonName, certRequest.Certificate.CommonName)

			if httpStatus == http.StatusOK {
				if orderDetails {
					return httpmock.NewJsonResponse(http.StatusOK, &digiCertRequestCertificateResponse{
						ID: 1234,
					})
				} else {
					return httpmock.NewJsonResponse(http.StatusOK, &digiCertRequestCertificateResponse{
						ID:               1234,
						CertificateID:    5678,
						CertificateChain: []certificateChain{{ee_cert}, {intermediate_cert}, {root_cert}},
					})
				}
			}
			return httpmock.NewJsonResponse(httpStatus, "{\"error_message\": \"Certificate profile with name \"Cert Profile\" doesn't exist\"}")
		},
	)
	certService := NewCertificateService()

	details, order, _ := certService.RequestCertificate(connection, pkcs10Request, domain.Product{
		OrganizationID: 1,
		HashAlgorithm:  "sha256",
		NameID:         "ssl_private_id",
	}, productOptionName, 300, &domain.ProductDetails{NameID: "ssl_private_id"})
	if httpStatus == http.StatusOK {
		if orderDetails {
			require.Equal(t, order.ID, "1234")
			require.Empty(t, order.CertificateID)
			require.Empty(t, details)
		} else {
			validateIssuanceCertificateDetails(t, details, "5678")
		}
	} else {
		require.Equal(t, details.Status, domain.CertificateStatusFailed)
		require.Equal(t, details.ErrorMessage, "failed to request certificate from DigiCert CA server: \"{\\\"error_message\\\": \\\"Certificate profile with name \\\"Cert Profile\\\" doesn't exist\\\"}\"")
		require.Empty(t, details.ID)
		require.Empty(t, details.Certificate)
		require.Empty(t, details.Chain)
	}
}

func testCheckCertificateData(t *testing.T, httpStatus int) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	certID := "CertID"
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

	httpmock.RegisterResponder("GET", fmt.Sprintf(downloadCertificateUri, certID),
		func(req *http.Request) (*http.Response, error) {
			if httpStatus == http.StatusOK {
				return httpmock.NewStringResponse(http.StatusOK, ee_cert+"\n"+intermediate_cert+"\n"+root_cert), nil
			}
			return httpmock.NewJsonResponse(httpStatus, "Some error")
		},
	)
	certificate := NewCertificateService()

	details, err := certificate.CheckCertificate(connection, certID)
	if httpStatus == http.StatusOK {
		validateIssuanceCertificateDetails(t, details, certID)
	} else {
		require.NotEmpty(t, err.Error())
	}
}

func testCheckOrderData(t *testing.T, httpStatus int) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	orderID := 1234
	certID := 5678
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

	httpmock.RegisterResponder("GET", fmt.Sprintf(orderCertificateUri, strconv.Itoa(orderID)),
		func(req *http.Request) (*http.Response, error) {
			if httpStatus == http.StatusOK {
				details := digiCertOrderDetails{
					ID:     orderID,
					Status: "issued",
					Certificate: &orderCertificate{
						ID: certID,
					},
				}
				return httpmock.NewJsonResponse(http.StatusOK, details)
			}
			return httpmock.NewJsonResponse(httpStatus, "{\"error_message\": \"Some error\"}")
		},
	)
	certificate := NewCertificateService()

	details, err := certificate.CheckOrder(connection, strconv.Itoa(orderID))
	if httpStatus == http.StatusOK {
		require.Equal(t, details.ID, strconv.Itoa(orderID))
		require.Equal(t, details.CertificateID, strconv.Itoa(certID))
		require.Equal(t, details.Status, domain.OrderStatusCompleted)
		require.Empty(t, details.ErrorMessage)
	} else {
		require.NotEmpty(t, err.Error())
	}
}

func testRetrieveCertificateData(t *testing.T, httpStatus int, cursor int, completed bool, includeExpired bool) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	nameID := "private_ssl_certificates"
	certID1 := 1234
	certID2 := 5678

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

	httpmock.RegisterResponder("GET", serverURL+"/order/certificate?filters[product_name_id]=private_ssl_certificates&filters[status]=issued&limit=2&offset=2&sort=order_id",
		func(req *http.Request) (*http.Response, error) {

			if httpStatus == http.StatusOK {
				if completed {
					return httpmock.NewJsonResponse(http.StatusOK, &digicertOrderDetailsSearchResponse{
						Page: page{
							Total:  cursor,
							Offset: cursor,
							Limit:  2,
						},
					})
				}
				digicertDetails := []digiCertOrderDetails{
					{
						ID:     1234,
						Status: "issued",
						Certificate: &orderCertificate{
							ID:        certID1,
							ValidTill: time.Now().AddDate(0, 0, 1).Format(digicertDateFormat),
						},
					},
					{
						ID:     1235,
						Status: "issued",
						Certificate: &orderCertificate{
							ID:        certID2,
							ValidTill: time.Now().AddDate(0, 0, -1).Format(digicertDateFormat),
						},
					},
				}
				return httpmock.NewJsonResponse(http.StatusOK, &digicertOrderDetailsSearchResponse{
					Orders: digicertDetails,
					Page: page{
						Total:  6,
						Offset: cursor,
						Limit:  2,
					},
				})
			}
			return httpmock.NewJsonResponse(httpStatus, "{\"error_message\": \"Some error\"}")
		},
	)

	httpmock.RegisterResponder("GET", fmt.Sprintf(downloadCertificateUri, strconv.Itoa(certID1)),
		func(req *http.Request) (*http.Response, error) {
			if httpStatus == http.StatusOK {
				return httpmock.NewStringResponse(http.StatusOK, ee_cert+"\n"+intermediate_cert+"\n"+root_cert), nil
			}
			return httpmock.NewJsonResponse(httpStatus, "Some error")
		},
	)

	httpmock.RegisterResponder("GET", fmt.Sprintf(downloadCertificateUri, strconv.Itoa(certID2)),
		func(req *http.Request) (*http.Response, error) {
			if httpStatus == http.StatusOK {
				return httpmock.NewStringResponse(http.StatusOK, ee_cert+"\n"+intermediate_cert+"\n"+root_cert), nil
			}
			return httpmock.NewJsonResponse(httpStatus, "Some error")
		},
	)

	certificate := NewCertificateService()

	option := domain.ImportOption{
		Name:        "Private SSL Certificates",
		Description: "Private SSL Certificates available for import",
		Settings: domain.ImportSettings{
			NameID: nameID,
		},
	}

	configuration := domain.ImportConfiguration{
		IncludeExpiredCertificates: includeExpired,
	}

	startCursor := strconv.Itoa(cursor)
	details, err := certificate.RetrieveCertificates(connection, option, configuration, startCursor, 2)
	if httpStatus == http.StatusOK {
		if completed {
			require.Equal(t, details.ImportStatus, domain.ImportStatusCompleted)
			require.Equal(t, details.LastProcessedCertificateID, startCursor)
			require.Empty(t, details.ImportCertificates)
		} else {
			require.Equal(t, details.ImportStatus, domain.ImportStatusUncompleted)
			require.Equal(t, details.LastProcessedCertificateID, strconv.Itoa(cursor+2))
			if !includeExpired {
				require.Equal(t, len(details.ImportCertificates), 1)
				validateCertificateDetails(t, details.ImportCertificates[0].Certificate, details.ImportCertificates[0].Chain, details.ImportCertificates[0].ID, strconv.Itoa(certID1))
			} else {
				require.Equal(t, len(details.ImportCertificates), 2)
				validateCertificateDetails(t, details.ImportCertificates[0].Certificate, details.ImportCertificates[0].Chain, details.ImportCertificates[0].ID, strconv.Itoa(certID1))
				validateCertificateDetails(t, details.ImportCertificates[1].Certificate, details.ImportCertificates[1].Chain, details.ImportCertificates[1].ID, strconv.Itoa(certID2))
			}
		}
	} else {
		assert.NotEmpty(t, err)
		assert.Nil(t, details)
	}
}

func validateIssuanceCertificateDetails(t *testing.T, details *domain.CertificateDetails, certID string) {
	validateCertificateDetails(t, details.Certificate, details.Chain, details.ID, certID)
	require.Equal(t, details.Status, domain.CertificateStatusIssued)
	require.Empty(t, details.ErrorMessage)
}

func validateCertificateDetails(t *testing.T, cert string, chain []string, certID string, expectedCertID string) {
	eePemBlock, _ := pem.Decode([]byte(ee_cert))
	intermPemBlock, _ := pem.Decode([]byte(intermediate_cert))
	rootPemBlock, _ := pem.Decode([]byte(root_cert))
	require.Equal(t, cert, base64.StdEncoding.EncodeToString(eePemBlock.Bytes))
	require.Equal(t, certID, expectedCertID)
	require.Equal(t, chain, []string{base64.StdEncoding.EncodeToString(intermPemBlock.Bytes), base64.StdEncoding.EncodeToString(rootPemBlock.Bytes)})
}
