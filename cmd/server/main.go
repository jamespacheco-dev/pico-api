package main

import (
	"log"
	"net/http"

	"github.com/jamespacheco-dev/pico-api/internal/api"
)

func main() {
	router := api.NewRouter()
	log.Println("Starting pico-api on :8080")
	log.Fatal(http.ListenAndServe(":8080", router))
}
