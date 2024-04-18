package service

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/venafi/digicert-ca-connector/internal/app/domain"
)

const (
	getOrganizationsUri = "/organization"
	getProductUri       = "/product"
)

type organization struct {
	ID       int    `json:"id"`
	Name     string `json:"name"`
	Status   string `json:"status"`
	IsActive bool   `json:"is_active"`
}

type getOrganizationsResponse struct {
	Organizations []organization `json:"organizations"`
}

type signatureHashTypes struct {
	AllowedHashTypes []hashType `json:"allowed_hash_types"`
	DefaultHashType  string     `json:"default_hash_type_id"`
}

type hashType struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type digiCertProductDetails struct {
	Hashes          signatureHashTypes `json:"signature_hash_types"`
	CertificateType string             `json:"type"`
	Name            string             `json:"name"`
	NameID          string             `json:"name_id"`
}

type getProductDetails struct {
	ProductDetails []digiCertProductDetails `json:"products"`
}

// Options ...
type Options struct {
}

// NewOptionsService will return a new webhook service
func NewOptionsService() *Options {
	return &Options{}
}

// GetOptions will retrieve product and import options from Certificate Authority
func (cs *Options) GetOptions(connection domain.Connection) ([]domain.ProductOption, []domain.ImportOption, error) {

	resp, err := executeRequest(connection, nil, getOrganizationsUri)

	if err != nil {
		return nil, nil, err
	}

	orgResponse := getOrganizationsResponse{}
	err = json.Unmarshal(resp.Body(), &orgResponse)
	if err != nil {
		return nil, nil, err
	}

	activeOrganizations := make([]int, 0)
	for _, org := range orgResponse.Organizations {
		if org.IsActive {
			activeOrganizations = append(activeOrganizations, org.ID)
		}
	}
	resp, err = executeRequest(connection, nil, getProductUri)
	if err != nil {
		return nil, nil, err
	}

	productResponse := getProductDetails{}
	err = json.Unmarshal(resp.Body(), &productResponse)
	if err != nil {
		return nil, nil, err
	}

	productOptions := make([]domain.ProductOption, 0)
	importOptions := make([]domain.ImportOption, 0)
	for _, product := range productResponse.ProductDetails {
		if product.CertificateType == "ssl_certificate" || product.CertificateType == "code_signing_certificate" {
			productType := domain.ProductTypeSsl
			if product.CertificateType == "code_signing_certificate" {
				productType = domain.ProductTypeCodeSign
			}
			hashes := make([]string, 0)
			for _, hash := range product.Hashes.AllowedHashTypes {
				hashes = append(hashes, hash.ID)
			}
			productOptions = append(productOptions, domain.ProductOption{
				Name:  product.Name,
				Types: []domain.ProductType{productType},
				Details: domain.ProductDetails{
					Hashes:               hashes,
					DefaultHashAlgorithm: product.Hashes.DefaultHashType,
					NameID:               product.NameID,
					Organizations:        activeOrganizations,
				},
			})

			importOptions = append(importOptions, domain.ImportOption{
				Name:        product.Name,
				Description: fmt.Sprintf("%s certificates will be available for import", product.CertificateType),
				Settings: domain.ImportSettings{
					NameID: product.NameID,
				},
			})
		}

	}
	return productOptions, importOptions, nil
}

// ValidateProduct will validate product against Certificate Authority
func (cs *Options) ValidateProduct(connection domain.Connection, name string, product domain.Product) ([]domain.ProductError, error) {

	options, _, err := cs.GetOptions(connection)
	if err != nil {
		return nil, err
	}

	var errors []domain.ProductError
	exist := false

	for _, option := range options {
		if option.Name == name {
			exist = true
			hashExist := false
			for _, hash := range option.Details.Hashes {
				if hash == product.HashAlgorithm {
					hashExist = true
					break
				}
			}
			orgExist := false
			for _, org := range option.Details.Organizations {
				if org == product.OrganizationID {
					orgExist = true
					break
				}
			}
			if !hashExist {
				errors = append(errors, domain.ProductError{
					AttributeName:  "hashAlgorithm",
					AttributeValue: product.HashAlgorithm,
				})
			}
			if !orgExist {
				errors = append(errors, domain.ProductError{
					AttributeName:  "organizationId",
					AttributeValue: strconv.Itoa(product.OrganizationID),
				})
			}
			break
		}
	}
	if !exist {
		errors = append(errors, domain.ProductError{
			AttributeName:  "name",
			AttributeValue: name,
		})
	}

	return errors, nil
}
