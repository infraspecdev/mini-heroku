package handlers

import (
	"encoding/json"
	"net/http"
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