package domain

type RevocationStatus string

const (
	RevocationStatusSubmitted RevocationStatus = "SUBMITTED"
	RevocationStatusFailed    RevocationStatus = "FAILED"
)

type CertificateRevocationData struct {
	SerialNumber            string `json:"serialNumber"`
	CaCertificateIdentifier string `json:"caCertificateIdentifier"`
	CaOrderIdentifier       string `json:"caOrderIdentifier"`
	Fingerprint             string `json:"fingerprint"`
	IssuerDN                string `json:"issuerDN"`
	CertificateContent      string `json:"certificateContent"`
}

type RevocationDetails struct {
	Status       RevocationStatus `json:"status"`
	ErrorMessage *string          `json:"errorMessage"`
}
