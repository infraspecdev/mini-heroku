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

	// Setup upload handler with Docker integration
	http.HandleFunc("/upload", func(w http.ResponseWriter, r *http.Request) {
		handlers.UploadHandlerWithDocker(w, r, dockerBuilder, dockerRunner)
	})

	fmt.Println("Controller listening on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
