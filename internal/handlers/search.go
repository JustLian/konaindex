package handlers

import (
	"encoding/json"
	"konaindex/internal/database"
	"konaindex/internal/models"
	"net/http"
	"strings"

	"github.com/lib/pq"
	"github.com/lucasb-eyer/go-colorful"
	"github.com/pgvector/pgvector-go"
)

type SearchRequest struct {
	Ratings      []string    `json:"ratings"`
	IncludeTags  []string    `json:"include_tags"`
	ExcludeTags  []string    `json:"exclude_tags"`
	TargetColors [][]float32 `json:"target_colors"`
	Limit        int         `json:"limit"`
	Page         int         `json:"page"`
}

func SearchHandler(w http.ResponseWriter, r *http.Request) {

	var req SearchRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// complex color search via intersections
	var postIDs []int
	if len(req.TargetColors) > 0 {

		var queryBuilder strings.Builder
		var args []interface{}

		for i, colorRGB := range req.TargetColors {
			if i > 0 {
				queryBuilder.WriteString(" INTERSECT ")
			}

			// converting RGB -> LAB
			c := colorful.Color{
				R: float64(colorRGB[0]) / 255.0,
				G: float64(colorRGB[1]) / 255.0,
				B: float64(colorRGB[2]) / 255.0,
			}
			l, a, b := c.Lab()

			vec := pgvector.NewVector([]float32{float32(l), float32(a), float32(b)})

			queryBuilder.WriteString(`(
				SELECT post_id FROM (
					SELECT post_id, weight, (color <-> ?) AS raw_dist
					FROM post_colors
					ORDER BY color <-> ? ASC
					LIMIT 300
				) AS fast_matches
				ORDER BY POWER(raw_dist, 2) / (weight + 0.01) ASC
				LIMIT 100
			)`)
			args = append(args, vec.String(), vec.String())
		}

		if err := database.DB.Raw(queryBuilder.String(), args...).Scan(&postIDs).Error; err != nil {
			WriteError(w, http.StatusInternalServerError, "Failed to search posts by color: "+err.Error())
			return
		}

		if len(postIDs) == 0 {
			WriteJSON(w, http.StatusOK, []models.Post{})
			return
		}

	}

	// creating query
	query := database.DB.Model(&models.Post{})

	// adding color search
	if len(postIDs) > 0 {
		query = query.Where("id IN ?", postIDs)
	}

	// adding rating filter
	if len(req.Ratings) > 0 {
		query = query.Where("rating IN ?", req.Ratings)
	}

	// including tags
	if len(req.IncludeTags) > 0 {
		query = query.Where("tags @> ?", pq.StringArray(req.IncludeTags))
	}

	// excluding tags
	if len(req.ExcludeTags) > 0 {
		query = query.Where("NOT tags && ?", pq.StringArray(req.ExcludeTags))
	}

	// pagination
	pageSize := req.Limit
	if pageSize <= 0 {
		pageSize = 20
	}

	page := req.Page
	if page < 1 {
		page = 1
	}

	offset := (page - 1) * pageSize

	var results []models.Post

	if err := query.Preload("Palette").Limit(pageSize).Offset(offset).Find(&results).Error; err != nil {
		WriteError(w, http.StatusInternalServerError, "Failed to search posts: "+err.Error())
		return
	}

	WriteJSON(w, http.StatusOK, results)
}
