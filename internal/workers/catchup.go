package workers

import (
	"fmt"
	"konaindex/internal/database"
	"konaindex/internal/models"
	"konaindex/internal/utils"
	"net/url"
	"time"
)

func StartCatchup(hardCapID int) {
	go func() {
		fmt.Println("[Catchup] Starting backfill...")

		for {

			var minID int
			database.DB.Model(&models.Post{}).Select("COALESCE(MIN(konachan_id), 0)").Scan(&minID)

			if minID == 0 {
				fmt.Println("[Catchup] Database is empty")
				time.Sleep(1 * time.Minute)
				continue
			}

			if minID <= hardCapID {
				fmt.Printf("[Catchup] Reached the historical cap (ID %d/%d). Backfill complete.\n", minID, hardCapID)
				return
			}

			// Fetching posts from the DB
			escaped := url.QueryEscape(fmt.Sprintf("id:<%d order:id_desc", minID))
			posts, err := utils.GetPosts(
				escaped,
				100, 0,
			)

			if err != nil || len(posts) == 0 {
				fmt.Printf("[Catchup] Error fetching posts: %v\n", err)
				time.Sleep(30 * time.Second)
				continue
			}

			// inserting posts
			inserted := utils.InsertPosts(posts)

			for _, p := range inserted {
				if uint(p.KonachanID) < uint(hardCapID) {
					fmt.Printf("[Catchup] Reached the historical cap (ID %d/%d). Backfill complete.\n", p.KonachanID, hardCapID)
					return
				}
				JobQueue <- p
			}

			fmt.Printf("[Catchup] Inserted %d posts\n", len(inserted))

			time.Sleep(18 * time.Second)
		}
	}()
}
