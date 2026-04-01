package client

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"mini-heroku/cli/signer"
)

// UploadPackage POSTs a signed tarball to the controller's /upload endpoint.
func UploadPackage(serverURL string, tarballReader io.Reader, appName, apiKey string) (*DeploymentResponse, error) {
	req, err := http.NewRequest(http.MethodPost, serverURL+UploadEndpoint, tarballReader)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}

	req.Header.Set(HeaderContentType, ContentTypeGzip)
	if appName != "" {
		req.Header.Set(HeaderAppName, appName)
	}
	attachAuth(req, apiKey)

	resp, err := (&http.Client{}).Do(req)
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

	var deployResp DeploymentResponse
	if err := json.Unmarshal(body, &deployResp); err != nil {
		return nil, fmt.Errorf("parsing response: %w", err)
	}
	return &deployResp, nil
}

func NewSignedRequest(method, url, apiKey string, body io.Reader) (*http.Request, error) {
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, err
	}
	attachAuth(req, apiKey)
	return req, nil
}

// attachAuth adds the API key header plus HMAC signing headers to req.
func attachAuth(req *http.Request, apiKey string) {
	if apiKey == "" {
		return
	}
	req.Header.Set(HeaderAPIKey, apiKey)
	for k, v := range signer.Headers(apiKey, req.Method, req.URL.Path) {
		req.Header.Set(k, v)
	}
}
