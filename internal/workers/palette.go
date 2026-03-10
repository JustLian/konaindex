package workers

import (
	"fmt"
	"image"
	"konaindex/internal/database"
	"konaindex/internal/models"
	"konaindex/internal/utils"
	"net/http"

	"github.com/pgvector/pgvector-go"
)

var JobQueue = make(chan models.Post, 300)

func logW(id int, message string, args ...interface{}) {
	fmt.Printf("[Worker %d] %s\n", id, fmt.Sprintf(message, args...))
}

func StartPool(numWorkers int) {
	for i := 0; i < numWorkers; i++ {
		go worker(i, JobQueue)
	}
	fmt.Printf("Started worker pool with %d active threads\n", numWorkers)
}

func worker(id int, jobs <-chan models.Post) {
	logW(id, "Worker started")
	for post := range jobs {

		logW(id, "Processing post %d", post.KonachanID)

		// requesting image
		resp, err := http.Get(post.PreviewURL)
		if err != nil {
			logW(id, "Failed to download image for post %d: %v", post.KonachanID, err)
			continue
		}

		// creating go image from bytes
		img, _, err := image.Decode(resp.Body)
		resp.Body.Close()
		if err != nil {
			logW(id, "Failed to decode image for post %d: %v", post.KonachanID, err)
			continue
		}

		// getting image info
		info, err := utils.GetImageInfo(img, 5)
		if err != nil {
			logW(id, "Failed to get image info for post %d: %v", post.KonachanID, err)
			continue
		}

		// deleting previous colors
		if err := database.DB.Where("post_id = ?", post.ID).Delete(&models.PostColor{}).Error; err != nil {
			logW(id, "Failed to delete old colors for post %d: %v", post.KonachanID, err)
			continue
		}

		// inserting colors
		colors := []models.PostColor{}
		for _, color := range info.Palette {
			colors = append(colors, models.PostColor{
				PostID: post.ID,
				Color:  pgvector.NewVector([]float32{float32(color.Color.L), float32(color.Color.A), float32(color.Color.B)}),
				Weight: color.Weight,
			})
		}
		if err := database.DB.Create(&colors).Error; err != nil {
			logW(id, "Failed to insert colors for post %d: %v", post.KonachanID, err)
			continue
		}

		// updating post with new data
		post.Temperature = info.Temperature
		database.DB.Save(&post)

		logW(id, "Finished processing post %d", post.KonachanID)
	}

	logW(id, "Worker exited")
}
