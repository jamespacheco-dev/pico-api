package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	GamesCreated = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "pico_games_created_total",
		Help: "Games created, by mode and digit length.",
	}, []string{"mode", "digits"})

	GamesCompleted = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "pico_games_completed_total",
		Help: "Games completed, by mode and difficulty.",
	}, []string{"mode", "difficulty"})

	GuessesPerGame = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "pico_guesses_per_game",
		Help:    "Number of guesses to complete a game.",
		Buckets: []float64{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 15, 20},
	}, []string{"mode", "difficulty"})

	Rollbacks = promauto.NewCounter(prometheus.CounterOpts{
		Name: "pico_rollbacks_total",
		Help: "Total rollback operations.",
	})

	ActiveGames = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "pico_active_games",
		Help: "Number of games currently in the store.",
	})

	AIGuessLatency = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "pico_ai_guess_latency_seconds",
		Help:    "Time spent computing the AI next guess (filtering + selection).",
		Buckets: prometheus.DefBuckets,
	}, []string{"difficulty"})

	HTTPRequests = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "pico_http_requests_total",
		Help: "HTTP requests, by method, route, and status code.",
	}, []string{"method", "route", "status"})

	HTTPDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "pico_http_request_duration_seconds",
		Help:    "HTTP request latency, by method and route.",
		Buckets: prometheus.DefBuckets,
	}, []string{"method", "route"})
)
