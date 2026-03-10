package workers

import (
	"fmt"
	"konaindex/internal/database"
	"konaindex/internal/models"
	"konaindex/internal/utils"
	"time"
)

func StartSync() {
	ticker := time.NewTicker(1 * time.Hour)

	runSync()

	go func() {
		for range ticker.C {
			runSync()
		}
	}()
}

func runSync() {

	// max id -> limit
	fmt.Println("[sync] waking up to check for new posts")
	var maxID int
	database.DB.Model(&models.Post{}).Select("COALESCE(MAX(konachan_id), 0)").Scan(&maxID)

	// fetching posts from the api
	posts, err := utils.GetPosts("", 100, 1)
	if err != nil {
		fmt.Println("[sync] error fetching posts:", err)
		return
	}
	fmt.Println("[sync] fetched", len(posts), "posts")

	// insertings posts
	var newPosts []utils.KonachanPost
	for _, post := range posts {
		if post.ID <= maxID {
			break
		}
		newPosts = append(newPosts, post)
	}

	insertedPosts := utils.InsertPosts(newPosts)
	fmt.Printf("[sync] inserted %d/%d new posts\n", len(insertedPosts), len(newPosts))

	// scheduling for processing
	for _, p := range insertedPosts {
		JobQueue <- p
	}
	fmt.Println("[sync] done")

}
