package handlers

import (
	"net/http"
	"strings"

	"mini-heroku/controller/internal/logger"
	"mini-heroku/controller/internal/store"
	"mini-heroku/controller/runner"
)

// URL pattern:  GET /apps/{appName}/logs
func LogsHandler(db store.StoreClient, dockerRunner runner.RunnerClient) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			sendError(w, http.StatusMethodNotAllowed, "Only GET allowed")
			return
		}

		// Extract appName from path: /apps/<appName>/logs
		appName := extractAppNameFromPath(r.URL.Path)
		if appName == "" {
			sendError(w, http.StatusBadRequest, "missing app name in path")
			return
		}

		appLog := logger.AppLogger(appName)

		// Look up the app in the database.
		project, err := db.GetByName(appName)
		if err != nil {
			appLog.Warn().Err(err).Msg("app not found")
			sendError(w, http.StatusNotFound, "app not found: "+appName)
			return
		}

		// Verify the container is actually running before attempting to tail.
		inspect, err := dockerRunner.ContainerInspect(r.Context(), project.ContainerID)
		if err != nil || !inspect.Running {
			sendError(w, http.StatusConflict, "container is not running for app: "+appName)
			return
		}

		// Confirm the RunnerClient can stream logs. The interface is extended in
		streamer, ok := dockerRunner.(runner.LogStreamer)
		if !ok {
			sendError(w, http.StatusInternalServerError, "log streaming not supported by runner")
			return
		}

		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("Cache-Control", "no-cache")
		w.WriteHeader(http.StatusOK)

		// Grab the Flusher so we can push bytes to the client incrementally.
		flusher, canFlush := w.(http.Flusher)

		appLog.Info().Str("container_id", project.ContainerID[:12]).Msg("streaming logs")

		if err := streamer.StreamLogs(r.Context(), project.ContainerID, w, flusher); err != nil {
			appLog.Warn().Err(err).Msg("log stream ended")
		}

		_ = canFlush // consumed inside StreamLogs; suppresses unused-variable lint
	}
}

func extractAppNameFromPath(path string) string {
	// Expected: /apps/<appName>/logs
	parts := strings.Split(strings.Trim(path, "/"), "/")
	if len(parts) == 3 && parts[0] == "apps" && parts[2] == "logs" {
		return parts[1]
	}
	return ""
}