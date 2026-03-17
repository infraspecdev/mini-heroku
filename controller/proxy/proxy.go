package proxy

import (
	"encoding/json"
	"fmt"
	"mini-heroku/controller/internal/logger"

	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
)

type Proxy struct {
	table *RouteTable
}

func NewProxy(table *RouteTable) *Proxy {
	return &Proxy{table: table}
}

func (p *Proxy) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	appName, err := extractAppName(r.Host)
	if err != nil {
		sendJSON(w, http.StatusBadRequest, "invalid host header: "+r.Host)
		return
	}

	targetURL, ok := p.table.Lookup(appName)
	if !ok {
		sendJSON(w, http.StatusNotFound, fmt.Sprintf("app %q not found", appName))
		return
	}

	target, err := url.Parse(targetURL)
	if err != nil {
		sendJSON(w, http.StatusInternalServerError, "bad target URL")
		return
	}

	appLog := logger.AppLogger(appName)
	appLog.Info().
		Str("method", r.Method).
		Str("host", r.Host).
		Str("target", targetURL).
		Msg("proxy request")

	rp := httputil.NewSingleHostReverseProxy(target)

	rp.ErrorHandler = func(w http.ResponseWriter, r *http.Request, err error) {
		appLog.Error().Err(err).Str("target", targetURL).Msg("error forwarding request")
		sendJSON(w, http.StatusBadGateway, "could not reach container")
	}

	rp.ServeHTTP(w, r)
}

func extractAppName(host string) (string, error) {
	host = strings.Split(host, ":")[0]

	if host == "" {
		return "", fmt.Errorf("empty host")
	}

	parts := strings.Split(host, ".")

	appName := parts[0]
	if appName == "" {
		return "", fmt.Errorf("empty app name in host: %s", host)
	}
	return appName, nil
}

func sendJSON(w http.ResponseWriter, status int, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]string{"error": msg})
}
