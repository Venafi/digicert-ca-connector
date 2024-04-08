package service

import (
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"regexp"
	"strconv"
	"time"

	"github.com/venafi/digicert-ca-connector/internal/app/domain"
	"go.uber.org/zap"
)

const (
	orderCertificateUri                     = "/order/certificate/%s"
	downloadCertificateUri                  = "/certificate/%s/download/format/pem_all"
	retrieveCertificatesUri                 = "/order/certificate?%sfilters[status]=issued&limit=%d&offset=%s&sort=order_id"
	retrieveCertificatesProductNameIdFilter = "filters[product_name_id]=%s&"
	digicertDateFormat                      = "2006-01-02"
)

type serverPlatform struct {
	ID int `json:"id"`
}

type certificate struct {
	CommonName           string         `json:"common_name"`
	DnsNames             []string       `json:"dns_names"`
	Csr                  string         `json:"csr"`
	ServerPlatform       serverPlatform `json:"server_platform"`
	SignatureHash        string         `json:"signature_hash"`
	CsProvisioningMethod string         `json:"cs_provisioning_method"`
}

type digicertOrganization struct {
	ID int `json:"id"`
}

type newCertificateRequestBody struct {
	Certificate          certificate          `json:"certificate"`
	Organization         digicertOrganization `json:"organization"`
	CustomExpirationDate string               `json:"custom_expiration_date"`
}

type certificateChain struct {
	Pem string `json:"pem"`
}

type digiCertRequestCertificateResponse struct {
	ID               int                `json:"id"`
	CertificateID    int                `json:"certificate_id"`
	CertificateChain []certificateChain `json:"certificate_chain"`
}

type orderCertificate struct {
	ID        int    `json:"id"`
	ValidTill string `json:"valid_till"`
}

type digiCertOrderDetails struct {
	ID          int               `json:"id"`
	Status      string            `json:"status"`
	Certificate *orderCertificate `json:"certificate"`
}

type page struct {
	Total  int `json:"total"`
	Limit  int `json:"limit"`
	Offset int `json:"offset"`
}

type digicertOrderDetailsSearchResponse struct {
	Orders []digiCertOrderDetails `json:"orders"`
	Page   page                   `json:"page"`
}

// Certificate service responsible for certificate related operations
type Certificate struct {
}

// NewCertificateService will return a new webhook certificate service
func NewCertificateService() *Certificate {
	return &Certificate{}
}

// RequestCertificate will request certificate from a Certificate Authority
func (cs *Certificate) RequestCertificate(connection domain.Connection, pkcs10Request string, product domain.Product, productOptionName string, validitySeconds int, productDetails *domain.ProductDetails) (*domain.CertificateDetails, *domain.OrderDetails, error) {

	pemBlock, _ := pem.Decode([]byte(pkcs10Request))
	csr, err := x509.ParseCertificateRequest(pemBlock.Bytes)
	if err != nil {
		return nil, nil, err
	}
	commonName := csr.Subject.CommonName
	if commonName == "" && len(csr.DNSNames) > 0 {
		commonName = csr.DNSNames[0]
	}

	if len(csr.DNSNames) == 0 {
		csr.DNSNames = append(csr.DNSNames, commonName)
	}

	re := regexp.MustCompile(`\r?\n`)
	pkcs10NoNewLines := re.ReplaceAllString(pkcs10Request, "")

	requestBody := newCertificateRequestBody{
		Certificate: certificate{
			CommonName: commonName,
			DnsNames:   csr.DNSNames,
			Csr:        pkcs10NoNewLines,
			ServerPlatform: serverPlatform{
				ID: -1,
			},
			SignatureHash: product.HashAlgorithm,
		},
		Organization: digicertOrganization{
			ID: product.OrganizationID,
		},
		CustomExpirationDate: time.Now().Add(time.Second * time.Duration(validitySeconds)).Format(digicertDateFormat),
	}

	resp, err := executeRequest(connection, requestBody, fmt.Sprintf(orderCertificateUri, productDetails.NameID))
	if err != nil {
		zap.L().Error(fmt.Sprintf("failed to request certificate from DigiCert CA using product name id: '%s'",
			productDetails.NameID), zap.Error(err))
		return &domain.CertificateDetails{
			Status:       domain.CertificateStatusFailed,
			ErrorMessage: fmt.Sprintf("failed to request certificate from DigiCert CA server: %s", err.Error()),
		}, nil, nil
	}

	digicertResponse := digiCertRequestCertificateResponse{}
	err = json.Unmarshal(resp.Body(), &digicertResponse)
	if err != nil {
		zap.L().Error("failed to unmarshal certificate data.", zap.Error(err))
		return &domain.CertificateDetails{
			Status:       domain.CertificateStatusFailed,
			ErrorMessage: fmt.Sprintf("failed request certificate from DigiCert CA server: %s", err.Error()),
		}, nil, nil
	}

	if digicertResponse.CertificateChain != nil || digicertResponse.CertificateID != 0 {
		certificateDetails := &domain.CertificateDetails{
			ID: strconv.Itoa(digicertResponse.CertificateID),
		}

		if digicertResponse.CertificateChain != nil {
			i := 0
			for _, cert := range digicertResponse.CertificateChain {
				if i == 0 {
					block, _ := pem.Decode([]byte(cert.Pem))
					certificateDetails.Certificate = base64.StdEncoding.EncodeToString(block.Bytes)
				} else {
					block, _ := pem.Decode([]byte(cert.Pem))
					certificateDetails.Chain = append(certificateDetails.Chain, base64.StdEncoding.EncodeToString(block.Bytes))
				}
				i++
			}
			certificateDetails.Status = domain.CertificateStatusIssued
		} else {
			certificateDetails.Status = domain.CertificateStatusRequested
		}
		return certificateDetails, nil, nil
	}

	orderDetails := &domain.OrderDetails{
		ID:     strconv.Itoa(digicertResponse.ID),
		Status: domain.OrderStatusProcessing,
	}
	if digicertResponse.CertificateID != 0 {
		orderDetails.CertificateID = strconv.Itoa(digicertResponse.CertificateID)
		orderDetails.Status = domain.OrderStatusCompleted
	}

	return nil, orderDetails, nil
}

// CheckOrder will check order details for submitted certificate request
func (cs *Certificate) CheckOrder(connection domain.Connection, id string) (*domain.OrderDetails, error) {

	resp, err := executeRequest(connection, nil, fmt.Sprintf(orderCertificateUri, id))
	if err != nil {
		return nil, err
	}

	digicertOrderDetails := digiCertOrderDetails{}
	err = json.Unmarshal(resp.Body(), &digicertOrderDetails)
	if err != nil {
		return nil, err
	}

	orderDetails := domain.OrderDetails{
		ID:     id,
		Status: domain.OrderStatusFailed,
	}
	if digicertOrderDetails.Status == "issued" {
		orderDetails.Status = domain.OrderStatusCompleted
		orderDetails.CertificateID = strconv.Itoa(digicertOrderDetails.Certificate.ID)
	} else if digicertOrderDetails.Status == "pending" || digicertOrderDetails.Status == "needs_approval" || digicertOrderDetails.Status == "processing" {
		orderDetails.Status = domain.OrderStatusProcessing
	}
	if digicertOrderDetails.Certificate != nil && digicertOrderDetails.Certificate.ID > 0 {
		orderDetails.CertificateID = strconv.Itoa(digicertOrderDetails.Certificate.ID)
	}

	return &orderDetails, nil
}

// CheckCertificate will check certificate details for submitted certificate request
func (cs *Certificate) CheckCertificate(connection domain.Connection, id string) (*domain.CertificateDetails, error) {

	resp, err := executeRequest(connection, nil, fmt.Sprintf(downloadCertificateUri, id))
	if err != nil {
		return nil, err
	}

	certDetails := domain.CertificateDetails{
		ID:     id,
		Status: domain.CertificateStatusRequested,
	}

	if resp.Body() != nil {
		cert, chain, err := parseCertificateData(resp.String())
		if err != nil {
			return nil, err
		}
		certDetails.Status = domain.CertificateStatusIssued
		certDetails.Certificate = cert
		certDetails.Chain = chain
	}

	return &certDetails, nil
}

// RetrieveCertificates will retrieve certificates available for import in TLSPC, from a Certificate Authority
func (cs *Certificate) RetrieveCertificates(connection domain.Connection, importOption domain.ImportOption, configuration domain.ImportConfiguration, startCursor string, batchSize int) (*domain.ImportDetails, error) {

	var filters = ""
	if importOption.Settings.NameID != "" {
		filters = fmt.Sprintf(retrieveCertificatesProductNameIdFilter, importOption.Settings.NameID)
	}
	resp, err := executeRequest(connection, nil, fmt.Sprintf(retrieveCertificatesUri, filters, batchSize, startCursor))
	if err != nil {
		return nil, err
	}

	orderDetailsSearchResponse := digicertOrderDetailsSearchResponse{}
	err = json.Unmarshal(resp.Body(), &orderDetailsSearchResponse)
	if err != nil {
		return nil, err
	}

	var lastOffset = orderDetailsSearchResponse.Page.Offset + len(orderDetailsSearchResponse.Orders)
	var status = domain.ImportStatusUncompleted
	if lastOffset == orderDetailsSearchResponse.Page.Total {
		status = domain.ImportStatusCompleted
	}

	var certificates []domain.ImportCertificate
	var now = time.Now()
	for _, order := range orderDetailsSearchResponse.Orders {
		dateValue, err := time.Parse(digicertDateFormat, order.Certificate.ValidTill)
		if err != nil {
			return nil, err
		}

		if dateValue.Before(now) && !configuration.IncludeExpiredCertificates {
			continue
		}

		resp, err = executeRequest(connection, nil, fmt.Sprintf("/certificate/%d/download/format/pem_all", order.Certificate.ID))
		if err != nil {
			return nil, err
		}

		if resp.Body() != nil {
			cert, chain, err := parseCertificateData(resp.String())
			if err != nil {
				continue
			}

			certificates = append(certificates, domain.ImportCertificate{
				ID:          strconv.Itoa(order.Certificate.ID),
				Certificate: cert,
				Chain:       chain,
			})
		}
	}

	return &domain.ImportDetails{
		ImportStatus:               status,
		LastProcessedCertificateID: strconv.Itoa(lastOffset),
		ImportCertificates:         certificates,
	}, nil
}

func parseCertificateData(pemData string) (string, []string, error) {

	certs, err := parseCertificatePEM([]byte(pemData))
	if err != nil {
		return "", nil, err
	}

	cert := base64.StdEncoding.EncodeToString(certs[0].Raw)

	var chain []string
	if len(certs) > 1 {
		for i := 1; i < len(certs); i++ {
			chain = append(chain, base64.StdEncoding.EncodeToString(certs[i].Raw))
		}
	}

	return cert, chain, nil
}

func parseCertificatePEM(certBytes []byte) ([]*x509.Certificate, error) {
	var certs []*x509.Certificate

	var block *pem.Block

	certDecoder := make([]byte, len(certBytes))
	copy(certDecoder, certBytes)

	for {
		// decode the tls certificate pem
		block, certDecoder = pem.Decode(certDecoder)
		if block == nil {
			break
		}

		// parse the tls certificate
		cert, err := x509.ParseCertificate(block.Bytes)
		if err != nil {
			return nil, fmt.Errorf("error parsing TLS certificate: %s", err)
		}
		certs = append(certs, cert)
	}

	if len(certs) == 0 {
		return nil, fmt.Errorf("failed to decode any certificate PEM block")
	}

	return certs, nil
}
