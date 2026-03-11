package handlers

import (
	"konaindex/internal/database"
	"konaindex/internal/models"
	"net/http"
)

func GetDBStatsHandler(w http.ResponseWriter, r *http.Request) {

	var count int64
	if err := database.DB.Model(&models.Post{}).Count(&count).Error; err != nil {
		WriteError(w, http.StatusInternalServerError, "Failed to count posts")
		return
	}

	WriteJSON(w, http.StatusOK, map[string]int64{"posts_indexed": count})
}
