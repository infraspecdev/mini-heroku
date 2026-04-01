package ratelimit

import (
	"encoding/json"
	"net/http"
	"sync"
	"time"

	"golang.org/x/time/rate"
)

// Config controls the token-bucket parameters.
type Config struct {
	// RequestsPerSecond is the sustained rate (tokens/second).
	RequestsPerSecond rate.Limit
	// Burst is the maximum instantaneous requests allowed.
	Burst int
}

// DefaultConfig is a sensible production default: 10 req/s, burst 20.
var DefaultConfig = Config{
	RequestsPerSecond: 10,
	Burst:             20,
}

type entry struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

// Middleware holds the per-IP limiter map.
type Middleware struct {
	mu      sync.Mutex
	clients map[string]*entry
	cfg     Config
}

// New creates a Middleware and starts the background cleanup goroutine.
func New(cfg Config) *Middleware {
	m := &Middleware{
		clients: make(map[string]*entry),
		cfg:     cfg,
	}
	go m.cleanup()
	return m
}

// Handler wraps next with per-IP rate limiting.
func (m *Middleware) Handler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip := clientIP(r)
		if !m.allow(ip) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusTooManyRequests)
			_ = json.NewEncoder(w).Encode(map[string]string{
				"status":  "error",
				"message": "rate limit exceeded — please slow down",
			})
			return
		}
		next.ServeHTTP(w, r)
	})
}

func (m *Middleware) allow(ip string) bool {
	m.mu.Lock()
	defer m.mu.Unlock()

	e, ok := m.clients[ip]
	if !ok {
		e = &entry{
			limiter: rate.NewLimiter(m.cfg.RequestsPerSecond, m.cfg.Burst),
		}
		m.clients[ip] = e
	}
	e.lastSeen = time.Now()
	return e.limiter.Allow()
}

// cleanup removes entries idle for more than 5 minutes.
func (m *Middleware) cleanup() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()
	for range ticker.C {
		m.mu.Lock()
		for ip, e := range m.clients {
			if time.Since(e.lastSeen) > 5*time.Minute {
				delete(m.clients, ip)
			}
		}
		m.mu.Unlock()
	}
}

// clientIP extracts the IP, preferring X-Forwarded-For.
func clientIP(r *http.Request) string {
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		return xff
	}
	return r.RemoteAddr
}
