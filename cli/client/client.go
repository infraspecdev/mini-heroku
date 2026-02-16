package client

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func UploadPackage(serverURL string, tarballReader io.Reader, appName string) (*DeploymentResponse, error) {
	// Create request
	req, err := http.NewRequest("POST", serverURL+UploadEndpoint, tarballReader)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}

	// Set headers
	req.Header.Set(HeaderContentType, ContentTypeGzip)
	if appName != "" {
		req.Header.Set(HeaderAppName, appName)
	}

	// Send request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("sending request: %w", err)
	}
	defer resp.Body.Close()

	// Parse response
	var deployResp DeploymentResponse
	if err := json.NewDecoder(resp.Body).Decode(&deployResp); err != nil {
		return nil, fmt.Errorf("parsing response: %w", err)
	}

	return &deployResp, nil
}
