package auth_test

import (
	"mini-heroku/controller/internal/auth"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRequireAPIKey_ValidKey(t *testing.T) {
	svc := auth.New("secret")
	handler := auth.RequireAPIKey(svc, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("X-API-Key", "secret")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestRequireAPIKey_MissingKey(t *testing.T) {
	svc := auth.New("secret")
	handler := auth.RequireAPIKey(svc, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", rec.Code)
	}
}

func TestRequireAPIKey_WrongKey(t *testing.T) {
	svc := auth.New("secret")
	handler := auth.RequireAPIKey(svc, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("X-API-Key", "bad-key")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", rec.Code)
	}
}

func TestRequireAPIKey_ResponseBody(t *testing.T) {
	svc := auth.New("secret")
	handler := auth.RequireAPIKey(svc, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	body := rec.Body.String()
	if body == "" {
		t.Error("expected non-empty JSON error body on 401")
	}
}
