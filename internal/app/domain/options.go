package domain

// ProductType type of the product
type ProductType string

const (
	ProductTypeSsl      ProductType = "SSL"
	ProductTypeCodeSign ProductType = "CODESIGN"
)

// Product contains needed product(issuance) data
type Product struct {
	OrganizationID int    `json:"organizationId"`
	HashAlgorithm  string `json:"hashAlgorithm"`
}

// ProductError represents attribute name and value for invalid product properties
type ProductError struct {
	AttributeName  string `json:"attributeName"`
	AttributeValue string `json:"attributeValue"`
}

// ProductDetails contains details related to available product option
type ProductDetails struct {
	Hashes               []string `json:"hashAlgorithms"`
	DefaultHashAlgorithm string   `json:"defaultHashAlgorithm"`
	NameID               string   `json:"nameId"`
	Organizations        []int    `json:"organizationIds"`
}

// ProductOption contains details related to available product(issuance) option
type ProductOption struct {
	Name    string         `json:"name"`
	Types   []ProductType  `json:"types"`
	Details ProductDetails `json:"productDetails"`
}

// ImportSettings contains settings related to available import option
type ImportSettings struct {
	NameID string `json:"nameId"`
}

// ImportOption contains details related to available import option
type ImportOption struct {
	Name        string         `json:"name"`
	Description string         `json:"description"`
	Settings    ImportSettings `json:"settings"`
}
