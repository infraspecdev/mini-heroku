package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"mini-heroku/controller/builder"
	"mini-heroku/controller/runner"
	"mini-heroku/controller/proxy"
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

	baseURL := os.Getenv("BASE_URL")
	if baseURL == "" {
		baseURL = "http://localhost"
	}

	sendSuccess(w, fmt.Sprintf("%s:%d", baseURL, HostPort), "App deployed successfully")
}

func UploadHandlerWithDocker(w http.ResponseWriter, r *http.Request, table *proxy.RouteTable, dockerBuilder builder.DockerClient, dockerRunner runner.RunnerClient) {
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

	log.Printf("[deploy] image built: %s", imageName)

	// Generate host port
	hostPort := runner.GenerateHostPort(appName)

	// Run container
	result, err := runner.RunContainer(dockerRunner, imageName, hostPort)
	if err != nil {
		sendError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to start container: %v", err))
		return
	}

	// Build container target URL
	targetURL := fmt.Sprintf("http://%s:8080", result.ContainerIP)

	// Register route in proxy
	table.Register(appName, targetURL)

	log.Printf("[deploy] route registered: %s -> %s", appName, targetURL)

	// Build public URL
	vmIP := os.Getenv("VM_PUBLIC_IP")
	if vmIP == "" {
		vmIP = "127.0.0.1"
	}

	publicURL := fmt.Sprintf("http://%s.%s.nip.io", appName, vmIP)

	// Success!
	sendSuccess(w, publicURL, "App deployed successfully")
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
