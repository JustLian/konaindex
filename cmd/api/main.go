package main

import (
	"fmt"
	"konaindex/config"
	"konaindex/internal/database"
	"konaindex/internal/handlers"
	"konaindex/internal/workers"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
)

func main() {
	cfg := config.MustLoad()

	// Connecting to the DB
	database.Connect(cfg.DatabaseURL)

	// Starting workers
	// workers.StartPool(cfg.WorkerCount)
	// workers.StartSync()
	// workers.StartCatchup(cfg.HistoricalCapID)
	workers.FastSaturateMissingMetadata()

	// Basic chi setup
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"http://localhost:*", "http://127.0.0.1:*", "https://konaindex.rian.moe"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	handlers.SetupRouters(r)

	port := ":" + cfg.ServerPort
	fmt.Printf("Starting KonaIndex API on http://localhost%s\n", port)

	err := http.ListenAndServe(port, r)
	if err != nil {
		log.Fatalf("Server crashed: %v", err)
	}
}
