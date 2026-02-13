package handlers

import (
    "bytes"
    "encoding/json"
    "net/http"
    "net/http/httptest"
    "testing"
)

func TestUploadHandler(t *testing.T) {
    tarballData := []byte("fake-gzip-data")
    
    req := httptest.NewRequest("POST", "/upload", bytes.NewReader(tarballData))
    req.Header.Set("Content-Type", "application/x-gzip")
    req.Header.Set("X-App-Name", "test-app")
    
    rec := httptest.NewRecorder()
    
    UploadHandler(rec, req)
    
    if rec.Code != http.StatusOK {
        t.Errorf("Expected 200 OK, got %d", rec.Code)
    }
    
    contentType := rec.Header().Get("Content-Type")
    if contentType != "application/json" {
        t.Errorf("Expected application/json, got %s", contentType)
    }
    
    var response DeploymentResponse
    if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
        t.Fatalf("Failed to parse response: %v", err)
    }
    

    if response.Status != "success" {
        t.Errorf("Expected success status, got %s", response.Status)
    }
}

func TestUploadHandlerRejectsNonPOST(t *testing.T) {
    req := httptest.NewRequest("GET", "/upload", nil)
    rec := httptest.NewRecorder()
    
    UploadHandler(rec, req)
    
    if rec.Code != http.StatusMethodNotAllowed {
        t.Errorf("Expected 405, got %d", rec.Code)
    }
}

func TestUploadHandlerRejectsWrongContentType(t *testing.T) {
    req := httptest.NewRequest("POST", "/upload", bytes.NewReader([]byte("data")))
    req.Header.Set("Content-Type", "text/plain")
    
    rec := httptest.NewRecorder()
    UploadHandler(rec, req)
    
    if rec.Code != http.StatusBadRequest {
        t.Errorf("Expected 400, got %d", rec.Code)
    }
}
