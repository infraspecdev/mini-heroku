package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"mini-heroku/controller/builder"
	"mini-heroku/controller/runner"
	"net/http"
	"os"
	"path/filepath"
)

const HostPort = 8888

func UploadHandler(w http.ResponseWriter, r *http.Request) {
	// Validate method
	if r.Method != http.MethodPost {
		sendError(w, http.StatusMethodNotAllowed, "Only POST allowed")
		return
	}

	// Validate content type
	if r.Header.Get(HeaderContentType) != ContentTypeGzip {
		sendError(w, http.StatusBadRequest, "Content-Type must be application/x-gzip")
		return
	}

	appName := r.Header.Get(HeaderAppName)
	if appName == "" {
		appName = "app-temp"
	}

	tempDir := os.TempDir()
	tarballPath := filepath.Join(tempDir, appName+".tar.gz")

	file, err := os.Create(tarballPath)
	if err != nil {
		sendError(w, http.StatusInternalServerError, "Failed to create temp file")
		return
	}
	defer file.Close()

	if _, err := io.Copy(file, r.Body); err != nil {
		sendError(w, http.StatusInternalServerError, "Failed to save upload")
		return
	}

	sendSuccess(w, "http://localhost:8888", "App deployed successfully")
}

func UploadHandlerWithDocker(w http.ResponseWriter, r *http.Request, dockerBuilder builder.DockerClient, dockerRunner runner.RunnerClient) {
	// Validate method
	if r.Method != http.MethodPost {
		sendError(w, http.StatusMethodNotAllowed, "Only POST allowed")
		return
	}

	// Validate content type
	contentType := r.Header.Get(HeaderContentType)
	if contentType != ContentTypeGzip {
		sendError(w, http.StatusBadRequest, "Content-Type must be application/x-gzip")
		return
	}

	// Get optional app name
	appName := r.Header.Get(HeaderAppName)
	if appName == "" {
		appName = generateRandomName()
	}

	// Save tarball to temp file for processing
	tempDir := os.TempDir()
	tarballPath := filepath.Join(tempDir, appName+".tar.gz")

	file, err := os.Create(tarballPath)
	if err != nil {
		sendError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to create temp file: %v", err))
		return
	}
	defer os.Remove(tarballPath) // Cleanup

	// Use io.Copy for streaming
	if _, err := io.Copy(file, r.Body); err != nil {
		file.Close()
		sendError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to save upload: %v", err))
		return
	}
	file.Close()

	// Read tarball for Docker processing
	tarballFile, err := os.Open(tarballPath)
	if err != nil {
		sendError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to read tarball: %v", err))
		return
	}
	defer tarballFile.Close()

	// Build Docker image directly from tarball
	imageName, err := builder.BuildImage(dockerBuilder, tarballFile, appName)
	if err != nil {
		sendError(w, http.StatusInternalServerError, fmt.Sprintf("Docker build failed: %v", err))
		return
	}

	// Run container
	appURL, err := runner.RunContainer(dockerRunner, imageName, HostPort)
	if err != nil {
		sendError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to start container: %v", err))
		return
	}

	// Success!
	sendSuccess(w, appURL, "App deployed successfully")
}

func sendSuccess(w http.ResponseWriter, appURL, message string) {
	w.Header().Set(HeaderContentType, ContentTypeJSON)
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(DeploymentResponse{
		Status:  StatusSuccess,
		AppURL:  appURL,
		Message: message,
	})
}

func sendError(w http.ResponseWriter, statusCode int, message string) {
	w.Header().Set(HeaderContentType, ContentTypeJSON)
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(DeploymentResponse{
		Status:  StatusError,
		Message: message,
	})
}

func generateRandomName() string {
	// Simple random name generator
	return fmt.Sprintf("app-%d", os.Getpid())
}
