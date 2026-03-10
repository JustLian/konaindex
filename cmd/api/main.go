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
)

func main() {
	cfg := config.MustLoad()

	// Connecting to the DB
	database.Connect(cfg.DatabaseURL)

	// Starting workers
	workers.StartPool(cfg.WorkerCount)
	workers.StartSync()
	workers.StartCatchup(cfg.HistoricalCapID)

	// Basic chi setup
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	handlers.SetupRouters(r)

	port := ":" + cfg.ServerPort
	fmt.Printf("Starting KonaIndex API on http://localhost%s\n", port)

	err := http.ListenAndServe(port, r)
	if err != nil {
		log.Fatalf("Server crashed: %v", err)
	}
}
