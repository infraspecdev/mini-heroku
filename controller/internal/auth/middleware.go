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
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(map[string]string{
				"status":  "error",
				"message": "unauthorized: invalid or missing API key",
			})
			return
		}
		next.ServeHTTP(w, r)
	})
}
