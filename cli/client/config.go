package client

// Configuration constants for CLI client
const (
	// Server configuration
	DefaultServerURL = "http://localhost:8080"
	UploadEndpoint   = "/upload"

	// HTTP Headers
	HeaderContentType = "Content-Type"
	HeaderAppName     = "App-Name"
	HeaderAPIKey      = "X-API-Key"

	// Content Types
	ContentTypeGzip = "application/x-gzip"
	ContentTypeJSON = "application/json"

	// Response statuses
	StatusSuccess = "success"
	StatusError   = "error"
)
