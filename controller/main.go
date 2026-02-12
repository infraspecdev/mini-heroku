package main

import (
	"fmt"
	"log"
	"mini-heroku/controller/handlers"
	"net/http"
)

func main() {
	http.HandleFunc("/upload", handlers.UploadHandler)

	fmt.Println("Controller listening on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
