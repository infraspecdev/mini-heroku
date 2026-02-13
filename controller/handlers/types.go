package handlers

type DeploymentResponse struct {
	Status  string `json:"status"`
	AppURL  string `json:"appUrl,omitempty"`
	Message string `json:"message"`
}
