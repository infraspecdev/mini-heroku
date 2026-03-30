package client

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

func UploadPackage(serverURL string, tarballReader io.Reader, appName string, apiKey string) (*DeploymentResponse, error) {

	// Create request
	req, err := http.NewRequest("POST", serverURL+UploadEndpoint, tarballReader)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}

	if apiKey != "" {
		req.Header.Set(HeaderAPIKey, apiKey)
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

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		var errResp DeploymentResponse
		if json.Unmarshal(body, &errResp) == nil && errResp.Message != "" {
			return nil, fmt.Errorf("controller returned %d: %s", resp.StatusCode, errResp.Message)
		}

		return nil, fmt.Errorf("controller returned %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}

	// Parse success response
	var deployResp DeploymentResponse
	if err := json.Unmarshal(body, &deployResp); err != nil {
		return nil, fmt.Errorf("parsing response: %w", err)
	}

	return &deployResp, nil
}
