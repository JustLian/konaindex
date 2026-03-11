package handlers

import (
	"konaindex/internal/database"
	"konaindex/internal/models"
	"net/http"
	"strconv"
)

func GetPostHandler(w http.ResponseWriter, r *http.Request) {
	idParam := r.URL.Query().Get("id")
	konachanIDParam := r.URL.Query().Get("konachan_id")

	if idParam == "" && konachanIDParam == "" {
		WriteError(w, http.StatusBadRequest, "Either 'id' or 'konachan_id' parameter is required")
		return
	}

	var post models.Post

	if idParam != "" {
		id, err := strconv.Atoi(idParam)
		if err != nil {
			WriteError(w, http.StatusBadRequest, "Invalid id format")
			return
		}
		if err := database.DB.Preload("Palette").First(&post, id).Error; err != nil {
			WriteError(w, http.StatusNotFound, "Post not found")
			return
		}
	} else {
		konachanID, err := strconv.Atoi(konachanIDParam)
		if err != nil {
			WriteError(w, http.StatusBadRequest, "Invalid konachan_id format")
			return
		}
		if err := database.DB.Preload("Palette").Where("konachan_id = ?", konachanID).First(&post).Error; err != nil {
			WriteError(w, http.StatusNotFound, "Post not found")
			return
		}
	}

	WriteJSON(w, http.StatusOK, post)
}
