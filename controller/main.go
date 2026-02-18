package main

import (
	"fmt"
	"log"
	"net/http"

	"mini-heroku/controller/builder"
	"mini-heroku/controller/handlers"
	"mini-heroku/controller/runner"
)

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

	// Create a new ServeMux
	mux := http.NewServeMux()

	// Register handlers
	mux.HandleFunc("/upload", func(w http.ResponseWriter, r *http.Request) {
		handlers.UploadHandlerWithDocker(w, r, dockerBuilder, dockerRunner)
	})
	mux.HandleFunc("/health", handlers.HealthHandler)

	// Wrap mux with custom 404 handler
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, pattern := mux.Handler(r)
		if pattern == "" {
			handlers.NotFoundHandler(w, r)
			return
		}
		mux.ServeHTTP(w, r)
	})

	// Create and start server
	server := &http.Server{
		Addr:    ":8080",
		Handler: handler,
	}

	fmt.Println("Controller listening on :8080")
	log.Fatal(server.ListenAndServe())
}
