package main

import (
	"log"
	"net/http"
	"time"

	"github.com/jamespacheco-dev/pico-api/internal/api"
	"github.com/jamespacheco-dev/pico-api/internal/metrics"
)

const (
	gameIdleTimeout  = 30 * time.Minute
	sweepInterval    = 5 * time.Minute
)

func main() {
	store := api.NewMemoryStore()

	go func() {
		ticker := time.NewTicker(sweepInterval)
		defer ticker.Stop()
		for range ticker.C {
			n := store.Sweep(gameIdleTimeout)
			if n > 0 {
				metrics.ActiveGames.Sub(float64(n))
				log.Printf("swept %d idle game(s)", n)
			}
		}
	}()

	router := api.NewRouter(store)
	log.Println("Starting pico-api on :8080")
	log.Fatal(http.ListenAndServe(":8080", router))
}
