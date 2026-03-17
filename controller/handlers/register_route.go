package handlers

import (
	"encoding/json"
	"mini-heroku/controller/internal/logger"
	"mini-heroku/controller/proxy"
	"net/http"
)

type RegisterRouteRequest struct {
	App    string `json:"app"`
	Target string `json:"target"`
}

func RegisterRouteHandler(table *proxy.RouteTable) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var req RegisterRouteRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "bad request: "+err.Error(), http.StatusBadRequest)
			return
		}

		if req.App == "" || req.Target == "" {
			http.Error(w, "app and target are required", http.StatusBadRequest)
			return
		}

		table.Register(req.App, req.Target)
		appLog := logger.AppLogger(req.App)
		appLog.Info().Str("target", req.Target).Msg("route registered")
		
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{
			"status": "ok",
			"app":    req.App,
			"target": req.Target,
		})
	}
}
