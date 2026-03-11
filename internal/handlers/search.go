package handlers

import (
	"encoding/json"
	"fmt"
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

	var postIDs []int
	if len(req.TargetColors) > 0 {
		var unions []string
		var havingClauses []string
		var scoreClauses []string
		var args []interface{}

		// processing each target color to build the Candidate Pool and Scorer
		for _, colorRGB := range req.TargetColors {
			c := colorful.Color{
				R: float64(colorRGB[0]) / 255.0,
				G: float64(colorRGB[1]) / 255.0,
				B: float64(colorRGB[2]) / 255.0,
			}
			l, a, b := c.Lab()
			vec := pgvector.NewVector([]float32{float32(l), float32(a), float32(b)})
			vecStr := vec.String()

			// 300 candidates
			unions = append(unions, "(SELECT post_id FROM post_colors ORDER BY color <-> ? LIMIT 300)")
			args = append(args, vecStr)

			// distance cap @ 25.0
			havingClauses = append(havingClauses, "MIN(pc.color <-> ?) < 25.0")
			args = append(args, vecStr)

			// calculating penalty
			scoreClauses = append(scoreClauses, "MIN(POWER(pc.color <-> ?, 2) / (pc.weight + 0.01))")
			args = append(args, vecStr)
		}

		// assembling
		rawQuery := fmt.Sprintf(`
			WITH candidate_pool AS (
				%s
			)
			SELECT pc.post_id
			FROM post_colors pc
			WHERE pc.post_id IN (SELECT post_id FROM candidate_pool)
			GROUP BY pc.post_id
			HAVING %s
			ORDER BY %s ASC
			LIMIT 1000
		`,
			strings.Join(unions, " UNION "),
			strings.Join(havingClauses, " AND "),
			fmt.Sprintf("GREATEST(%s)", strings.Join(scoreClauses, ", ")),
		)

		if err := database.DB.Raw(rawQuery, args...).Scan(&postIDs).Error; err != nil {
			WriteError(w, http.StatusInternalServerError, "Color search failed: "+err.Error())
			return
		}

		// early stop if no results
		if len(postIDs) == 0 {
			WriteJSON(w, http.StatusOK, []models.Post{})
			return
		}
	}

	// querying
	query := database.DB.Model(&models.Post{})

	if len(postIDs) > 0 {
		// converting to comma-separated string
		var idStrings []string
		for _, id := range postIDs {
			idStrings = append(idStrings, fmt.Sprintf("%d", id))
		}
		idList := strings.Join(idStrings, ",")

		query = query.Where("id IN ?", postIDs)

		orderQuery := fmt.Sprintf("array_position(ARRAY[%s]::int[], posts.id)", idList)
		query = query.Order(orderQuery)
	}

	if len(req.Ratings) > 0 {
		query = query.Where("rating IN ?", req.Ratings)
	}

	if len(req.IncludeTags) > 0 {
		query = query.Where("tags @> ?", pq.StringArray(req.IncludeTags))
	}

	if len(req.ExcludeTags) > 0 {
		query = query.Where("NOT tags && ?", pq.StringArray(req.ExcludeTags))
	}

	// pagination Setup
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
		WriteError(w, http.StatusInternalServerError, "Database retrieval failed: "+err.Error())
		return
	}

	WriteJSON(w, http.StatusOK, results)
}
