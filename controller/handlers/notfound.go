package handlers

import (
	"encoding/json"
	"net/http"
)

func NotFoundHandler(w http.ResponseWriter, r *http.Request) {
	response := map[string]string{
		"error": "endpoint not found",
		"path":  r.URL.Path,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusNotFound)
	json.NewEncoder(w).Encode(response)
}
