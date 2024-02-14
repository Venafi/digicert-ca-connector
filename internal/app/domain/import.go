package domain

// ImportConfiguration contains import configuration
type ImportConfiguration struct {
	IncludeExpiredCertificates bool `json:"includeExpiredCertificates"`
}

// ImportStatus status for the import.
type ImportStatus string

const (
	// ImportStatusCompleted represents completed status for the import.
	ImportStatusCompleted ImportStatus = "COMPLETED"
	// ImportStatusUncompleted represents uncompleted status for the import.
	ImportStatusUncompleted ImportStatus = "UNCOMPLETED"
)

// ImportCertificate contains details for imported certificate
type ImportCertificate struct {
	ID          string   `json:"id"`
	Certificate string   `json:"certificate"`
	Chain       []string `json:"chain"`
}

// ImportDetails contains details for the import
type ImportDetails struct {
	ImportStatus               ImportStatus        `json:"status"`
	LastProcessedCertificateID string              `json:"lastProcessedCertificateId"`
	ImportCertificates         []ImportCertificate `json:"certificates"`
}
