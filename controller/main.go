package main

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"mini-heroku/controller/builder"
	"mini-heroku/controller/handlers"
	"mini-heroku/controller/internal/store"
	"mini-heroku/controller/proxy"
	"mini-heroku/controller/runner"
)

func reconcile(s *store.Store, runnerClient runner.RunnerClient, table *proxy.RouteTable) {
	ctx := context.Background()

	projects, err := s.ListAll()
	if err != nil {
		log.Printf("reconcile: db.ListAll failed: %v", err)
		return
	}

	log.Printf("reconciliation started: count=%d", len(projects))
	registered := 0

	for i := range projects {
		p := &projects[i]
		prefix := fmt.Sprintf("[%s]", p.Name)

		inspect, inspectErr := runnerClient.ContainerInspect(ctx, p.ContainerID)

		if inspectErr == nil && inspect.Running {
			// Container is already running
			p.ContainerIP = inspect.IPAddress
			log.Printf("%s container already running: id=%s", prefix, p.ContainerID[:12])

		} else if inspectErr == nil && !inspect.Running {
			// Container exists but is stopped
			log.Printf("%s container stopped — restarting: id=%s", prefix, p.ContainerID[:12])

			if err := runnerClient.ContainerStart(ctx, p.ContainerID); err != nil {
				log.Printf("%s container restart failed: %v", prefix, err)
				_ = s.UpdateStatus(p.Name, "error")
				continue
			}
			// Re-inspect to get fresh IP after start
			freshInspect, err := runnerClient.ContainerInspect(ctx, p.ContainerID)
			if err != nil {
				log.Printf("%s inspect after restart failed: %v", prefix, err)
				_ = s.UpdateStatus(p.Name, "error")
				continue
			}

			p.ContainerIP = freshInspect.IPAddress
			p.Status = "running"
			_ = s.Upsert(p)

			log.Printf("%s container restarted: id=%s", prefix, p.ContainerID[:12])

		} else {
			log.Printf("%s container not found — creating new container: old_id=%s", prefix, p.ContainerID[:12])

			var hostPortInt int
			fmt.Sscanf(p.HostPort, "%d", &hostPortInt)

			result, err := runner.RunContainer(runnerClient, p.ImageName, hostPortInt)
			if err != nil {
				log.Printf("%s container creation failed: %v", prefix, err)
				_ = s.UpdateStatus(p.Name, "error")
				continue
			}

			p.ContainerID = result.ContainerID
			p.ContainerIP = result.ContainerIP
			p.Status = "running"
			_ = s.Upsert(p)

			log.Printf("%s new container created: id=%s", prefix, result.ContainerID[:12])
		}

		targetURL := fmt.Sprintf("http://localhost:%s", p.HostPort)
		table.Register(p.Name, targetURL)
		registered++

		log.Printf("%s route registered: target=%s", prefix, targetURL)
	}

	log.Printf("reconciliation complete: registered=%d", registered)
}

func main() {
	// Initialize real Docker clients
	dockerBuilder, err := builder.NewRealDockerClient()
	if err != nil {
		log.Fatalf("Failed to create Docker builder client: %v", err)
	}

	dockerRunner, err := runner.NewRealRunnerClient()
	if err != nil {
		log.Fatalf("Failed to create Docker runner client: %v", err)
	}

	// Initialize SQLite store (mini.db in working directory)
	db, err := store.NewStore("mini.db")
	if err != nil {
		log.Fatalf("store init: %v", err)
	}

	// Create shared RouteTable
	table := proxy.NewRouteTable()

	// Create a new ServeMux
	mux := http.NewServeMux()

	// Reconcile: rebuild proxy routes from DB before accepting requests
	reconcile(db, dockerRunner, table)

	// Register handlers
	mux.HandleFunc("/upload", func(w http.ResponseWriter, r *http.Request) {
		handlers.UploadHandlerWithDocker(w, r, table, dockerBuilder, dockerRunner, db)
	})

	mux.HandleFunc("/health", handlers.HealthHandler)

	// Route registration endpoint
	mux.HandleFunc("/register-route", handlers.RegisterRouteHandler(table))

	// Wrap mux with custom 404 handler
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, pattern := mux.Handler(r)
		if pattern == "" {
			handlers.NotFoundHandler(w, r)
			return
		}
		mux.ServeHTTP(w, r)
	})

	// Start reverse proxy in separate goroutine
	p := proxy.NewProxy(table)

	go func() {
		log.Println("[proxy] listening on :80")
		err := http.ListenAndServe(":80", p)
		if err != nil {
			log.Fatalf("[proxy] error: %v", err)
		}
	}()

	// Create and start server
	server := &http.Server{
		Addr:    ":8080",
		Handler: handler,
	}

	fmt.Println("Controller listening on :8080")
	log.Fatal(server.ListenAndServe())
}
