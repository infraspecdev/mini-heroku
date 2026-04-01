package ratelimit_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"mini-heroku/controller/internal/ratelimit"
)

func okHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
}

func TestRateLimiter_AllowsRequestsUnderBurst(t *testing.T) {
	m := ratelimit.New(ratelimit.Config{RequestsPerSecond: 100, Burst: 50})
	h := m.Handler(okHandler())

	for i := 0; i < 10; i++ {
		req := httptest.NewRequest(http.MethodGet, "/health", nil)
		req.RemoteAddr = "10.0.0.1:9000"
		rr := httptest.NewRecorder()
		h.ServeHTTP(rr, req)

		if rr.Code != http.StatusOK {
			t.Errorf("request %d: got %d, want %d", i+1, rr.Code, http.StatusOK)
		}
	}
}

func TestRateLimiter_BlocksAfterBurstExhausted(t *testing.T) {
	// Burst=1, rate=1/s — second immediate request must be blocked.
	m := ratelimit.New(ratelimit.Config{RequestsPerSecond: 1, Burst: 1})
	h := m.Handler(okHandler())

	send := func(ip string) int {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.RemoteAddr = ip
		rr := httptest.NewRecorder()
		h.ServeHTTP(rr, req)
		return rr.Code
	}

	if code := send("5.5.5.5:1"); code != http.StatusOK {
		t.Fatalf("first request got %d, want 200", code)
	}
	if code := send("5.5.5.5:1"); code != http.StatusTooManyRequests {
		t.Fatalf("second immediate request got %d, want 429", code)
	}
}

func TestRateLimiter_DifferentIPsHaveSeparateBuckets(t *testing.T) {
	m := ratelimit.New(ratelimit.Config{RequestsPerSecond: 1, Burst: 1})
	h := m.Handler(okHandler())

	ips := []string{"1.1.1.1:0", "2.2.2.2:0", "3.3.3.3:0"}
	for _, ip := range ips {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.RemoteAddr = ip
		rr := httptest.NewRecorder()
		h.ServeHTTP(rr, req)

		if rr.Code != http.StatusOK {
			t.Errorf("IP %s first request got %d, want 200", ip, rr.Code)
		}
	}
}

func TestRateLimiter_RespectsXForwardedFor(t *testing.T) {
	m := ratelimit.New(ratelimit.Config{RequestsPerSecond: 1, Burst: 1})
	h := m.Handler(okHandler())

	req1 := httptest.NewRequest(http.MethodGet, "/", nil)
	req1.Header.Set("X-Forwarded-For", "9.9.9.9")
	rr1 := httptest.NewRecorder()
	h.ServeHTTP(rr1, req1)

	req2 := httptest.NewRequest(http.MethodGet, "/", nil)
	req2.Header.Set("X-Forwarded-For", "9.9.9.9")
	rr2 := httptest.NewRecorder()
	h.ServeHTTP(rr2, req2)

	if rr2.Code != http.StatusTooManyRequests {
		t.Errorf("second request from same forwarded IP: got %d, want 429", rr2.Code)
	}
}

func TestRateLimiter_ResponseBodyOnBlock(t *testing.T) {
	m := ratelimit.New(ratelimit.Config{RequestsPerSecond: 1, Burst: 1})
	h := m.Handler(okHandler())

	send := func() *httptest.ResponseRecorder {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.RemoteAddr = "7.7.7.7:1"
		rr := httptest.NewRecorder()
		h.ServeHTTP(rr, req)
		return rr
	}

	send() // consume burst
	rr := send()

	if rr.Code != http.StatusTooManyRequests {
		t.Fatalf("got %d, want 429", rr.Code)
	}
	if ct := rr.Header().Get("Content-Type"); ct != "application/json" {
		t.Errorf("Content-Type = %q, want application/json", ct)
	}
}
