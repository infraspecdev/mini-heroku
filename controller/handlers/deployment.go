package handlers

import (
	"encoding/json"
	"io"
	"net/http"
	"os"
	"path/filepath"
)

func UploadHandler(w http.ResponseWriter, r *http.Request) {
	// Validate method
	if r.Method != http.MethodPost {
		sendError(w, http.StatusMethodNotAllowed, "Only POST allowed")
		return
	}

	// Validate content type
	if r.Header.Get("Content-Type") != "application/x-gzip" {
		sendError(w, http.StatusBadRequest, "Content-Type must be application/x-gzip")
		return
	}

	appName := r.Header.Get("App-Name")
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

func sendSuccess(w http.ResponseWriter, appURL, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(DeploymentResponse{
		Status:  StatusSuccess,
		AppURL:  appURL,
		Message: message,
	})
}

func sendError(w http.ResponseWriter, statusCode int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(DeploymentResponse{
		Status:  StatusError,
		Message: message,
	})
}