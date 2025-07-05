package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"time"
)

// Top-level response structure
type RSS2JSONResponse struct {
	Status string `json:"status"`
	Feed   Feed   `json:"feed"`
	Items  []Item `json:"items"`
}

// Feed structure
type Feed struct {
	URL         string `json:"url"`
	Title       string `json:"title"`
	Link        string `json:"link"`
	Author      string `json:"author"`
	Description string `json:"description"`
	Image       string `json:"image"`
}

// Item structure (for each RSS/blog post)
type Item struct {
	Title       string   `json:"title"`
	PubDate     string   `json:"pubDate"` // Can be parsed to time.Time if needed
	Link        string   `json:"link"`
	GUID        string   `json:"guid"`
	Author      string   `json:"author"`
	Thumbnail   string   `json:"thumbnail"`
	Description string   `json:"description"`
	Content     string   `json:"content"`
	Enclosure   struct{} `json:"enclosure"` // Empty object, so an empty struct is fine
	Categories  []string `json:"categories"`
}

// httpClient is a custom HTTP client with a timeout for best practices.
var httpClient = &http.Client{
	Timeout: time.Second * 10, // 10-second timeout for the request
}

func main() {
	apiURL := "https://api.rss2json.com/v1/api.json?rss_url=https://medium.com/feed/@lav.nya.verma"

	// Create a new GET request
	req, err := http.NewRequest(http.MethodGet, apiURL, nil)
	if err != nil {
		log.Fatalf("Error creating request: %v", err)
	}

	// Set a User-Agent header (good practice)
	req.Header.Set("User-Agent", "Go-RSS2JSON-Client/1.0")

	// Perform the request
	resp, err := httpClient.Do(req)
	if err != nil {
		log.Fatalf("Error making HTTP request: %v", err)
	}
	defer resp.Body.Close() // Crucial: Close the response body

	// Check the HTTP status code
	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := ioutil.ReadAll(resp.Body) // Read body for more context in error
		log.Fatalf("External API returned non-OK status: %d, body: %s", resp.StatusCode, string(bodyBytes))
	}

	// Read the response body
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatalf("Error reading response body: %v", err)
	}

	// Unmarshal the JSON response into our RSS2JSONResponse struct
	var rssData RSS2JSONResponse
	err = json.Unmarshal(body, &rssData)
	if err != nil {
		log.Fatalf("Error unmarshaling JSON: %v, body: %s", err, string(body))
	}

	// Now you can work with the parsed data
	fmt.Printf("Status: %s\n", rssData.Status)
	fmt.Printf("Feed Title: %s\n", rssData.Feed.Title)
	fmt.Printf("Number of items: %d\n", len(rssData.Items))

	if len(rssData.Items) > 0 {
		fmt.Println("\nFirst Item Details:")
		fmt.Printf("  Title: %s\n", rssData.Items[0].Title)
		fmt.Printf("  Author: %s\n", rssData.Items[0].Author)
		fmt.Printf("  Published Date: %s\n", rssData.Items[0].PubDate)
		fmt.Printf("  Link: %s\n", rssData.Items[0].Link)
		fmt.Printf("  Categories: %v\n", rssData.Items[0].Categories)
		// You can print Description and Content, but they can be very long.
		// fmt.Printf("  Description: %s\n", rssData.Items[0].Description)
		// fmt.Printf("  Content: %s\n", rssData.Items[0].Content)
	}
}
