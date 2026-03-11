package handlers

import (
	"github.com/go-chi/chi/v5"
)

func SetupRouters(r chi.Router) {
	r.Route("/api", func(r chi.Router) {
		r.Post("/search", SearchHandler)
		r.Get("/stats", GetDBStatsHandler)
		r.Get("/post", GetPostHandler)
	})

	r.Get("/", GetRootHandler)
}
