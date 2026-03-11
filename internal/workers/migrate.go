package workers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"konaindex/internal/database"
	"konaindex/internal/models"
)

type KonachanMigrationResponse struct {
	ID       int `json:"id"`
	Width    int `json:"width"`
	Height   int `json:"height"`
	Score    int `json:"score"`
	FileSize int `json:"file_size"`
}

func FastSaturateMissingMetadata() {
	var maxID, minID int

	// 1. Find our boundaries
	database.DB.Model(&models.Post{}).Select("COALESCE(MAX(konachan_id), 0)").Scan(&maxID)
	database.DB.Model(&models.Post{}).Select("COALESCE(MIN(konachan_id), 0)").Scan(&minID)

	if maxID == 0 {
		fmt.Println("[Migration] Database is empty. Nothing to migrate.")
		return
	}

	fmt.Printf("[Migration] Starting batch backfill from ID %d down to %d\n", maxID, minID)

	// Set cursor just above maxID so the first query includes the maxID itself
	cursor := maxID + 1
	totalUpdated := 0

	for cursor >= minID {
		url := fmt.Sprintf("https://konachan.net/post.json?limit=100&tags=id:<%d", cursor)
		req, err := http.NewRequest("GET", url, nil)
		if err != nil {
			fmt.Printf("[Migration] Error creating request at cursor %d: %v\n", cursor, err)
			time.Sleep(5 * time.Second)
			continue
		}
		req.Header.Set("User-Agent", "KonaIndex/1.0 (https://github.com/JustLian/konaindex)")
		req.Header.Set("Accept", "application/json")

		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			fmt.Printf("[Migration] Error making request at cursor %d: %v\n", cursor, err)
			time.Sleep(5 * time.Second)
			continue
		}

		var kResp []KonachanMigrationResponse
		if err := json.NewDecoder(resp.Body).Decode(&kResp); err != nil || len(kResp) == 0 {
			fmt.Printf("[Migration] No more posts found below ID %d. Exiting.\n", cursor)
			resp.Body.Close()
			break
		}

		batchUpdated := 0
		lowestIDInBatch := cursor // Track this to advance the pagination

		// 2. Process the batch
		for _, kp := range kResp {
			if kp.ID < lowestIDInBatch {
				lowestIDInBatch = kp.ID
			}

			// We only run the UPDATE if the row exists AND is actually missing data (width = 0).
			// This makes the database operation incredibly fast and safe.
			result := database.DB.Model(&models.Post{}).
				Where("konachan_id = ? AND width IS NULL", kp.ID).
				Updates(map[string]interface{}{
					"width":     kp.Width,
					"height":    kp.Height,
					"score":     kp.Score,
					"file_size": kp.FileSize,
				})

			if result.RowsAffected > 0 {
				batchUpdated++
				totalUpdated++
			}
		}

		fmt.Printf("[Migration] Scanned 100 posts (Cursor: %d). Fetched: %d. Updated: %d. Total Fixed: %d\n", cursor, len(kResp), batchUpdated, totalUpdated)
		resp.Body.Close()

		// 3. Move the cursor down for the next loop
		cursor = lowestIDInBatch

		// Play nice with the API limits
		time.Sleep(3 * time.Second)
	}

	fmt.Printf("[Migration] FINISHED! Successfully saturated %d records.\n", totalUpdated)
}
