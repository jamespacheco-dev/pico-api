package main

import (
	"log"
	"net/http"

	"github.com/jamespacheco-dev/pico-api/internal/api"
)

func main() {
	store := api.NewMemoryStore()
	router := api.NewRouter(store)
	log.Println("Starting pico-api on :8080")
	log.Fatal(http.ListenAndServe(":8080", router))
}
