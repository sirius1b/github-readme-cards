package client

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	. "github.com/sirius1b/github-readme-cards/internal"
	. "github.com/sirius1b/github-readme-cards/util"
)

var (
	httpClient = &http.Client{
		Timeout: time.Second * 5, // 5-second timeout for the request
	}
	apiURL = "https://api.rss2json.com/v1/api.json?rss_url=https://medium.com/feed/@"
)

func GetUserData(user string, db map[string]FeedResponse) (RSS2JSONResponse, error) {
	log.Printf("getUserData called for user: %s", user)
	data, ok := db[user]

	if !ok || time.Since(data.UpdatedAt) > time.Second*time.Duration(Validity) {
		log.Printf("Cache miss or expired for user: %s. Fetching from API.", user)
		resp, err := httpClient.Get(apiURL + user)
		if err != nil {
			log.Printf("Error fetching data from API for user %s: %v", user, err)
			return RSS2JSONResponse{}, err
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			log.Printf("Non-OK HTTP status: %d for user %s", resp.StatusCode, user)
			return RSS2JSONResponse{}, err
		}

		var feedResponse RSS2JSONResponse
		if err := json.NewDecoder(resp.Body).Decode(&feedResponse); err != nil {
			log.Printf("Error decoding JSON for user %s: %v", user, err)
			return RSS2JSONResponse{}, err
		}
		db[user] = FeedResponse{
			Rss:       feedResponse,
			UpdatedAt: time.Now(),
		}
		data = db[user]
	} else {
		log.Printf("Cache hit for user: %s", user)
	}

	return data.Rss, nil
}
