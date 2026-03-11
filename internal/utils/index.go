package utils

import (
	"encoding/json"
	"fmt"
	"konaindex/internal/database"
	"konaindex/internal/models"
	"net/http"
	"strings"
)

type KonachanPost struct {
	ID         int    `json:"id"`
	Tags       string `json:"tags"`
	FileURL    string `json:"jpeg_url"`
	PreviewURL string `json:"preview_url"`
	Rating     string `json:"rating"`
	Width      int    `json:"width"`
	Height     int    `json:"height"`
	Score      int    `json:"score"`
	FileSize   int    `json:"file_size"`
}

// Fetches a slice of KonachanPost from the API
func GetPosts(tags string, limit int, page int) ([]KonachanPost, error) {

	// constructing the URL
	url := "https://konachan.net/post.json?"
	var elems []string
	if tags != "" {
		elems = append(elems, "tags="+tags)
	}
	if limit != 0 {
		elems = append(elems, fmt.Sprintf("limit=%d", limit))
	}
	if page != 0 {
		elems = append(elems, fmt.Sprintf("page=%d", page))
	}

	url += strings.Join(elems, "&")

	// making the request
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "KonaIndex/1.0 (https://github.com/JustLian/konaindex)")
	req.Header.Set("Accept", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var kPosts []KonachanPost
	if err := json.NewDecoder(resp.Body).Decode(&kPosts); err != nil {
		return nil, err
	}

	return kPosts, nil
}

// Inserts a slice of KonachanPost into the database
func InsertPosts(kPosts []KonachanPost) []models.Post {

	inserted := []models.Post{}

	for _, kp := range kPosts {

		post := models.Post{
			KonachanID: kp.ID,
			Tags:       strings.Split(kp.Tags, " "),
			ImageURL:   kp.FileURL,
			PreviewURL: kp.PreviewURL,
			Rating:     kp.Rating,
			Width:      kp.Width,
			Height:     kp.Height,
			Score:      kp.Score,
			FileSize:   kp.FileSize,
		}

		result := database.DB.Where(models.Post{KonachanID: kp.ID}).FirstOrCreate(&post)

		if result.RowsAffected > 0 {
			inserted = append(inserted, post)
		}
	}

	return inserted

}
