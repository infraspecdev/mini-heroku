package auth

import (
	"encoding/json"
	"net/http"
)

const APIKeyHeader = "X-API-Key"

func RequireAPIKey(svc *AuthService, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		key := r.Header.Get(APIKeyHeader)

		if !svc.Validate(key) {
			writeJSON(w, http.StatusUnauthorized, "unauthorized: invalid or missing API key")
			return
		}

		if err := ValidateTimestamp(r); err != nil {
			writeJSON(w, http.StatusUnauthorized, "unauthorized: "+err.Error())
			return
		}

		if err := ValidateHMAC(r, svc.secret()); err != nil {
			writeJSON(w, http.StatusUnauthorized, "unauthorized: "+err.Error())
			return
		}

		next.ServeHTTP(w, r)
	})
}

func writeJSON(w http.ResponseWriter, status int, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]string{
		"status":  "error",
		"message": msg,
	})
}
