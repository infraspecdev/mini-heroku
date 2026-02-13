package client

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestUploadPackage(t *testing.T) {
	// Setup: Create mock server
	var receivedMethod string
	var receivedContentType string
	var receivedAppName string
	var receivedBody []byte

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedMethod = r.Method
		receivedContentType = r.Header.Get(HeaderContentType)
		receivedAppName = r.Header.Get(HeaderAppName)
		receivedBody, _ = io.ReadAll(r.Body)

		// Respond with success
		w.Header().Set(HeaderContentType, ContentTypeJSON)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{
            "status": "success",
            "appUrl": "http://localhost:8888",
            "message": "Deployed"
        }`))
	}))
	defer server.Close()

	// Execute: Upload dummy tarball
	tarballData := []byte("fake-tarball-data")
	response, err := UploadPackage(server.URL, bytes.NewReader(tarballData), "test-app")

	// Assert: Request was correct
	if err != nil {
		t.Fatalf("Upload failed: %v", err)
	}

	if receivedMethod != "POST" {
		t.Errorf("Expected POST, got %s", receivedMethod)
	}

	if receivedContentType != ContentTypeGzip {
		t.Errorf("Expected %s, got %s", ContentTypeGzip, receivedContentType)
	}

	if receivedAppName != "test-app" {
		t.Errorf("Expected test-app, got %s", receivedAppName)
	}

	if !bytes.Equal(receivedBody, tarballData) {
		t.Error("Body data mismatch")
	}

	// Assert: Response was parsed
	if response.Status != StatusSuccess {
		t.Errorf("Expected %s, got %s", StatusSuccess, response.Status)
	}

	if response.AppURL != "http://localhost:8888" {
		t.Errorf("Expected localhost:8888, got %s", response.AppURL)
	}
}

func TestUploadPackageWithoutAppName(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		appName := r.Header.Get(HeaderAppName)
		if appName != "" {
			t.Error("Expected no app name header")
		}

		w.Header().Set(HeaderContentType, ContentTypeJSON)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status": "success", "appUrl": "http://localhost:8888", "message": "OK"}`))
	}))
	defer server.Close()

	tarballData := []byte("data")
	response, err := UploadPackage(server.URL, bytes.NewReader(tarballData), "")

	if err != nil {
		t.Fatalf("Upload failed: %v", err)
	}

	if response.Status != StatusSuccess {
		t.Errorf("Expected success, got %s", response.Status)
	}
}
