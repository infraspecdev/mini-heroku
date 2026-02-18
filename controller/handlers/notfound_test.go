package handlers

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestNotFoundHandler(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/invalid", nil)
	w := httptest.NewRecorder()

	NotFoundHandler(w, req)

	res := w.Result()
	defer res.Body.Close()

	if res.StatusCode != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", res.StatusCode)
	}
}
