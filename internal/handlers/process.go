package handlers

import (
	"encoding/json"
	"konaindex/internal/database"
	"konaindex/internal/models"
	"konaindex/internal/workers"
	"net/http"
)

type ProcessRequest struct {
	KonachanIDs []int `json:"konachan_ids"`
}

func ProcessPostsHandler(w http.ResponseWriter, r *http.Request) {

	var req ProcessRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		WriteError(w, http.StatusBadRequest, "Invalid JSON payload")
		return
	}

	// Fetching posts from DB
	var posts []models.Post
	database.DB.Where("konachan_id IN ?", req.KonachanIDs).Find(&posts)

	// Scheduling tasks
	for _, post := range posts {
		workers.JobQueue <- post
	}

	WriteJSON(w, http.StatusAccepted, map[string]interface{}{
		"message":       "Jobs added to queue",
		"requested_ids": req.KonachanIDs,
		"found_posts":   len(posts),
		"queued":        len(posts),
	})

}
