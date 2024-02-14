package domain

// CertificateStatus status for the submitted certificate request.
type CertificateStatus string

const (
	// CertificateStatusPending represents pending status for submitted certificate request.
	CertificateStatusPending CertificateStatus = "PENDING"
	// CertificateStatusRequested represents requested status for submitted certificate request.
	CertificateStatusRequested CertificateStatus = "REQUESTED"
	// CertificateStatusIssued represents issued status for submitted certificate request.
	CertificateStatusIssued CertificateStatus = "ISSUED"
	// CertificateStatusFailed represents failed status for submitted certificate request.
	CertificateStatusFailed CertificateStatus = "FAILED"
)

// CertificateDetails contains certificate details for the submitted certificate request to a Certificate Authority
type CertificateDetails struct {
	ID           string            `json:"id"`
	Status       CertificateStatus `json:"status"`
	Certificate  string            `json:"certificate"`
	Chain        []string          `json:"chain"`
	ErrorMessage string            `json:"errorMessage"`
}
