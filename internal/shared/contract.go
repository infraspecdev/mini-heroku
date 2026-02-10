package shared

const (
	// Controller (API server)
	ControllerPort = 8080

	// API endpoints
	UploadEndpoint = "/upload"

	// Headers / content types
	ContentTypeTGZ = "application/x-gzip"
)

type UploadResponse struct {
	AppURL  string `json:"appUrl,omitempty"`
	Message string `json:"message,omitempty"`
}
