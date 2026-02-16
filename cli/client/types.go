package client

// DeploymentResponse represents the server's response to an upload
type DeploymentResponse struct {
	Status  string `json:"status"`  // "success" or "error"
	AppURL  string `json:"appUrl"`  // Only present on success
	Message string `json:"message"` // Description of result
}
