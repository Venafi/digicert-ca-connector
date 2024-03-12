package domain

// Connection contains needed configuration and credentials to connect to a Certificate Authority
type Connection struct {
	Configuration Configuration `json:"configuration"`
	Credentials   Credentials   `json:"credentials"`
}

// Configuration contains needed configuration for connection to a Certificate Authority
type Configuration struct {
	ServerURL string `json:"serverUrl"`
}

// Credentials contains needed credentials to authenticate against a Certificate Authority
type Credentials struct {
	ApiKey string `json:"apiKey"`
}
