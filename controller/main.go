package main

import (
	"context"
	"fmt"
	"net/http"
	"os"

	"mini-heroku/controller/builder"
	"mini-heroku/controller/handlers"
	"mini-heroku/controller/internal/auth"
	"mini-heroku/controller/internal/logger"
	"mini-heroku/controller/internal/ratelimit"
	"mini-heroku/controller/internal/store"
	"mini-heroku/controller/proxy"
	"mini-heroku/controller/runner"
)

func reconcile(s *store.Store, runnerClient runner.RunnerClient, table *proxy.RouteTable) {
	ctx := context.Background()
	projects, err := s.ListAll()
	if err != nil {
		logger.Log.Error().Err(err).Msg("reconcile: db.ListAll failed")
		return
	}

	logger.Log.Info().Int("count", len(projects)).Msg("reconciliation started")
	registered := 0

	for i := range projects {
		p := &projects[i]
		appLog := logger.AppLogger(p.Name)

		inspect, inspectErr := runnerClient.ContainerInspect(ctx, p.ContainerID)

		if inspectErr == nil && inspect.Running {
			p.ContainerIP = inspect.IPAddress
			appLog.Info().Str("container_id", p.ContainerID[:12]).Msg("container already running")
		} else if inspectErr == nil && !inspect.Running {
			appLog.Warn().Str("container_id", p.ContainerID[:12]).Msg("container stopped — restarting")
			if err := runnerClient.ContainerStart(ctx, p.ContainerID); err != nil {
				appLog.Error().Err(err).Msg("container restart failed")
				_ = s.UpdateStatus(p.Name, "error")
				continue
			}
			freshInspect, err := runnerClient.ContainerInspect(ctx, p.ContainerID)
			if err != nil {
				appLog.Error().Err(err).Msg("inspect after restart failed")
				_ = s.UpdateStatus(p.Name, "error")
				continue
			}
			p.ContainerIP = freshInspect.IPAddress
			p.Status = "running"
			_ = s.Upsert(p)
			appLog.Info().Str("container_id", p.ContainerID[:12]).Msg("container restarted")
		} else {
			appLog.Warn().Str("container_id", p.ContainerID[:12]).Msg("container not found — creating new")
			var hostPortInt int
			fmt.Sscanf(p.HostPort, "%d", &hostPortInt)

			result, err := runner.RunContainer(runnerClient, p.ImageName, hostPortInt)
			if err != nil {
				appLog.Error().Err(err).Msg("container creation failed")
				_ = s.UpdateStatus(p.Name, "error")
				continue
			}
			p.ContainerID = result.ContainerID
			p.ContainerIP = result.ContainerIP
			p.Status = "running"
			_ = s.Upsert(p)
			appLog.Info().Str("container_id", result.ContainerID[:12]).Msg("new container created")
		}

		targetURL := fmt.Sprintf("http://localhost:%s", p.HostPort)
		table.Register(p.Name, targetURL)
		registered++
		appLog.Info().Str("target", targetURL).Msg("route registered")
	}

	logger.Log.Info().Int("registered", registered).Msg("reconciliation complete")
}

func main() {
	logger.Init()

	apiKey := os.Getenv("API_KEY")
	if apiKey == "" {
		logger.Log.Warn().Msg("API_KEY env var not set — all protected routes will reject requests")
	}
	authSvc := auth.New(apiKey)

	dockerBuilder, err := builder.NewRealDockerClient()
	if err != nil {
		logger.Log.Fatal().Err(err).Msg("failed to create Docker builder client")
	}

	dockerRunner, err := runner.NewRealRunnerClient()
	if err != nil {
		logger.Log.Fatal().Err(err).Msg("failed to create Docker runner client")
	}

	db, err := store.NewStore("/opt/mini-heroku/data/mini.db")
	if err != nil {
		logger.Log.Fatal().Err(err).Msg("store init failed")
	}
	logger.Log.Info().Str("path", "/opt/mini-heroku/data/mini.db").Msg("database ready")

	table := proxy.NewRouteTable()
	reconcile(db, dockerRunner, table)

	// Rate limiter applied globally to the controller API.
	rl := ratelimit.New(ratelimit.DefaultConfig)

	mux := http.NewServeMux()

	mux.Handle("/upload", auth.RequireAPIKey(authSvc, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handlers.UploadHandlerWithDocker(w, r, table, dockerBuilder, dockerRunner, db)
	})))

	mux.HandleFunc("/health", handlers.HealthHandler)

	mux.Handle("/register-route", auth.RequireAPIKey(authSvc,
		http.HandlerFunc(handlers.RegisterRouteHandler(table)),
	))

	mux.Handle("/apps/", auth.RequireAPIKey(authSvc,
		http.HandlerFunc(handlers.LogsHandler(db, dockerRunner)),
	))

	// Custom 404 + global rate limiter wrapping the entire mux.
	handler := rl.Handler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, pattern := mux.Handler(r)
		if pattern == "" {
			handlers.NotFoundHandler(w, r)
			return
		}
		mux.ServeHTTP(w, r)
	}))

	p := proxy.NewProxy(table)
	go func() {
		logger.Log.Info().Str("addr", ":80").Msg("proxy listening")
		if err := http.ListenAndServe(":80", p); err != nil {
			logger.Log.Fatal().Err(err).Msg("proxy server failed")
		}
	}()

	logger.Log.Info().Str("addr", ":8080").Msg("controller listening")
	if err := (&http.Server{Addr: ":8080", Handler: handler}).ListenAndServe(); err != nil {
		logger.Log.Fatal().Err(err).Msg("http server failed")
	}
}
