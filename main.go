package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/net/html"
)

type FeedResponse struct {
	Rss       RSS2JSONResponse
	UpdatedAt time.Time
}

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
	PubDate     string   `json:"pubDate"`
	Link        string   `json:"link"`
	GUID        string   `json:"guid"`
	Author      string   `json:"author"`
	Thumbnail   string   `json:"thumbnail"`
	Description string   `json:"description"`
	Content     string   `json:"content"`
	Enclosure   struct{} `json:"enclosure"`
	Categories  []string `json:"categories"`
}

// ParsedPubDate parses the PubDate string to time.Time
func (i *Item) ParsedPubDate() (time.Time, error) {
	return time.Parse("2006-01-02 15:04:05", i.PubDate)
}

var (
	httpClient = &http.Client{
		Timeout: time.Second * 5, // 5-second timeout for the request
	}
	db     = make(map[string]FeedResponse)
	apiURL = "https://api.rss2json.com/v1/api.json?rss_url=https://medium.com/feed/@"
)

const (
	validity        = 600
	defaultMaxCount = 10
	defaultType     = Latest
	defaultTheme    = GithubLight
)

type QueryType int

const (
	Latest QueryType = iota
)

type ColorTheme int

const (
	SolarizedLight ColorTheme = iota
	GithubLight
	RosePineDawn
	QuietLight
	Dracula
	GruvBox
	Nord
	OneDark
	Monokai
	TokyoNight
)

// ThemeColors holds color values for a theme
type ThemeColors struct {
	Primary    string
	Secondary  string
	Accent     string
	Background string
	Text       string
}

// ThemeColorMap maps ColorTheme constants to their ThemeColors
var ThemeColorMap = map[ColorTheme]ThemeColors{
	SolarizedLight: {
		Primary:    "#268bd2",
		Secondary:  "#2aa198",
		Accent:     "#b58900",
		Background: "#fdf6e3",
		Text:       "#657b83",
	},
	GithubLight: {
		Primary:    "#0969da",
		Secondary:  "#6e7781",
		Accent:     "#d4a72c",
		Background: "#ffffff",
		Text:       "#24292f",
	},
	RosePineDawn: {
		Primary:    "#b4637a",
		Secondary:  "#ea9d34",
		Accent:     "#56949f",
		Background: "#faf4ed",
		Text:       "#575279",
	},
	QuietLight: {
		Primary:    "#6c6c6c",
		Secondary:  "#b3b3b3",
		Accent:     "#ffab70",
		Background: "#f5f5f5",
		Text:       "#333333",
	},
	Dracula: {
		Primary:    "#bd93f9",
		Secondary:  "#ff79c6",
		Accent:     "#50fa7b",
		Background: "#282a36",
		Text:       "#f8f8f2",
	},
	GruvBox: {
		Primary:    "#fabd2f",
		Secondary:  "#b8bb26",
		Accent:     "#fe8019",
		Background: "#282828",
		Text:       "#ebdbb2",
	},
	Nord: {
		Primary:    "#5e81ac",
		Secondary:  "#88c0d0",
		Accent:     "#a3be8c",
		Background: "#2e3440",
		Text:       "#d8dee9",
	},
	OneDark: {
		Primary:    "#61afef",
		Secondary:  "#c678dd",
		Accent:     "#e5c07b",
		Background: "#282c34",
		Text:       "#abb2bf",
	},
	Monokai: {
		Primary:    "#f92672",
		Secondary:  "#a6e22e",
		Accent:     "#fd971f",
		Background: "#272822",
		Text:       "#f8f8f2",
	},
	TokyoNight: {
		Primary:    "#7aa2f7",
		Secondary:  "#bb9af7",
		Accent:     "#7dcfff",
		Background: "#1a1b26",
		Text:       "#c0caf5",
	},
}

// ------------------------------------------------------------------------------------------------------

func queryFromString(query string) QueryType {
	log.Printf("queryFromString called with query: %s", query)
	switch query {
	case "latest":
		return Latest
	default:
		return defaultType // Default to Latest if unknown
	}
}

func countFromString(countStr string) int {
	log.Printf("countFromString called with countStr: %s", countStr)
	if countStr == "" {
		return defaultMaxCount
	}
	count, err := strconv.Atoi(countStr)
	if err != nil || count <= 0 {
		log.Printf("Invalid count value: %s, defaulting to %d", countStr, defaultMaxCount)
		return defaultMaxCount
	}
	return count
}

func themeFromString(themeStr string) ColorTheme {
	log.Printf("themeFromString called with themeStr: %s", themeStr)
	switch strings.ToLower(themeStr) {
	case "solarizedlight":
		return SolarizedLight
	case "githublight":
		return GithubLight
	case "rosepinedawn":
		return RosePineDawn
	case "quietlight":
		return QuietLight
	case "dracula":
		return Dracula
	case "gruvbox":
		return GruvBox
	case "nord":
		return Nord
	case "onedark":
		return OneDark
	case "monokai":
		return Monokai
	case "tokyonight":
		return TokyoNight
	default:
		log.Printf("Unknown theme: %s, defaulting to GithubLight", themeStr)
		return defaultTheme // Default to GithubLight if unknown
	}
}

func getUserData(user string) (RSS2JSONResponse, error) {
	log.Printf("getUserData called for user: %s", user)
	data, ok := db[user]

	if !ok || time.Since(data.UpdatedAt) > time.Second*time.Duration(validity) {
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

func parseIt(rss RSS2JSONResponse, queryType QueryType, count int, color ThemeColors) (string, error) {
	log.Printf("parseIt called with queryType: %v", queryType)
	articles := rss.Items
	if len(articles) > count {
		articles = articles[:count]
	}
	log.Printf("Number of articles to process: %d", len(articles))

	// Prepare header fields
	feed := rss.Feed
	avatar := feed.Image
	if avatar == "" {
		avatar = "https://cdn-icons-png.flaticon.com/512/5968/5968885.png"
	}
	author := feed.Author
	if author == "" {
		author = feed.Title
	}
	title := feed.Title
	if title == "" {
		title = "Medium Feed"
	}

	svg := `<svg width="800" height="` + strconv.Itoa(len(articles)*100+(len(articles)-1)*15+110+50) + `" xmlns="http://www.w3.org/2000/svg" font-family="Segoe UI, sans-serif">`
	svg += `
  <defs>
	<linearGradient id="grad" x1="0" y1="0" x2="1" y2="1">
	  <stop offset="0%" stop-color="` + color.Background + `"/>
	  <stop offset="100%" stop-color="` + color.Secondary + `"/>
	</linearGradient>
	<filter id="cardShadow" x="-10%" y="-10%" width="120%" height="120%">
	  <feDropShadow dx="0" dy="1" stdDeviation="2" flood-color="` + color.Secondary + `"/>
	</filter>
  </defs>
  <rect width="100%" height="100%" fill="url(#grad)" />
  <!-- Header -->
  <g transform="translate(30, 30)">
	<circle cx="30" cy="30" r="30" fill="` + color.Background + `" />
	<image href="` + avatar + `" x="0" y="0" width="60" height="60" clip-path="circle(30px at 30px 30px)" />
	<text x="80" y="28" font-size="24" font-weight="600" fill="` + color.Primary + `">` + escapeXML(author) + `</text>
	<text x="80" y="50" font-size="16" fill="` + color.Text + `">📝 Latest from Medium</text>
  </g>
  <!-- Cards Container -->
  <g transform="translate(30, 110)">
`
	cardYOffset := 0
	cardHeight := 100
	cardSpacing := 15
	for i := range articles {
		// Format date
		article := articles[i]
		pubDate := article.PubDate
		t, err := article.ParsedPubDate()
		if err == nil {
			pubDate = t.Format("Jan 2, 2006")
		}
		// Category
		category := ""
		if len(article.Categories) > 0 {
			category = "#" + article.Categories[0]
		}
		// Description (trimmed)
		formatted, error := GetPlainTextFromHTML(article.Description)
		desc := ""
		if error == nil {
			desc = trimString(formatted, 90)
			log.Printf("Formatted description for article %s: %s", article.Title, desc)
		}
		svg += `
	<g transform="translate(0, ` + intToString(cardYOffset) + `)" filter="url(#cardShadow)">
	  <rect width="740" height="100" rx="16" ry="16" fill="` + color.Background + `" opacity="0.95"/>
	  <a href="` + article.Link + `">
		<text x="20" y="28" font-size="16" font-weight="600" fill="` + color.Primary + `">` + escapeXML(trimString(article.Title, 85)) + `</text>
	  </a>
	  <text x="20" y="47" font-size="12" fill="` + color.Secondary + `" font-family="monospace">📅 ` + escapeXML(pubDate)
		if category != "" {
			svg += ` • ` + escapeXML(category)
		}

		if desc != "" {
			svg += `</text>
				<text x="20" y="67" font-size="13" fill="` + color.Text + `">` + escapeXML(desc) + `</text>
				</g>
			`
		} else {
			svg += `</text>
					</g>
				`
		}

		cardYOffset += cardHeight + cardSpacing

	}
	svg += `
  </g>
</svg>
`
	log.Printf("SVG generated with %d articles", len(articles))
	return svg, nil
}

func GetPlainTextFromHTML(htmlString string) (string, error) {
	doc, err := html.Parse(strings.NewReader(htmlString))
	if err != nil {
		return "", fmt.Errorf("failed to parse HTML: %w", err)
	}

	var buf bytes.Buffer
	var f func(*html.Node)
	f = func(n *html.Node) {
		// Skip script and style tags and their content
		if n.Type == html.ElementNode && (n.Data == "script" || n.Data == "style") {
			return
		}

		if n.Type == html.TextNode {
			// Trim whitespace from text nodes.
			// html.TextNode can contain newlines and multiple spaces.
			trimmedText := strings.TrimSpace(n.Data)
			if trimmedText != "" {
				// Add a space before appending if the buffer is not empty
				// and the last character is not already a space/newline.
				if buf.Len() > 0 && buf.Bytes()[buf.Len()-1] != ' ' && buf.Bytes()[buf.Len()-1] != '\n' {
					buf.WriteByte(' ')
				}
				buf.WriteString(trimmedText)
			}
		}

		// Add a newline for block-level elements for better readability
		if n.Type == html.ElementNode {
			switch n.Data {
			case "p", "div", "h1", "h2", "h3", "h4", "h5", "h6", "li", "br":
				if buf.Len() > 0 && buf.Bytes()[buf.Len()-1] != '\n' { // Avoid double newlines
					buf.WriteByte('\n')
				}
			}
		}

		for c := n.FirstChild; c != nil; c = c.NextSibling {
			f(c)
		}

		// Add a newline after block-level elements too, for separation
		if n.Type == html.ElementNode {
			switch n.Data {
			case "p", "div", "h1", "h2", "h3", "h4", "h5", "h6", "li":
				if buf.Len() > 0 && buf.Bytes()[buf.Len()-1] != '\n' { // Avoid double newlines
					buf.WriteByte('\n')
				}
			}
		}
	}
	f(doc)

	// Final cleanup: remove excessive newlines and leading/trailing whitespace
	cleanedText := strings.ReplaceAll(buf.String(), "\n\n", "\n")
	cleanedText = strings.TrimSpace(cleanedText)
	return cleanedText, nil
}

// intToString is a helper to convert int to string
func intToString(i int) string {
	return fmt.Sprintf("%d", i)
}

// trimString trims a string to the specified length and adds ellipsis if needed
func trimString(s string, maxLen int) string {
	if len([]rune(s)) <= maxLen {
		return s
	}
	runes := []rune(s)
	if maxLen > 3 {
		return string(runes[:maxLen-3]) + "..."
	}
	if maxLen > 0 {
		return string(runes[:maxLen])
	}
	return ""
}

// escapeXML escapes special XML characters in a string
func escapeXML(s string) string {
	var buf []rune
	for _, r := range s {
		switch r {
		case '&':
			buf = append(buf, []rune("&amp;")...)
		case '<':
			buf = append(buf, []rune("&lt;")...)
		case '>':
			buf = append(buf, []rune("&gt;")...)
		case '"':
			buf = append(buf, []rune("&quot;")...)
		case '\'':
			buf = append(buf, []rune("&apos;")...)
		default:
			buf = append(buf, r)
		}
	}
	return string(buf)
}

func setupRouter() *gin.Engine {
	log.Println("setupRouter called")
	r := gin.Default()

	r.GET("/ping", func(c *gin.Context) {
		log.Println("GET /ping called")
		c.String(http.StatusOK, "pong")
	})

	r.GET("/user/:name", func(c *gin.Context) {
		user := c.Params.ByName("name")
		log.Printf("GET /user/%s called", user)
		userData, err := getUserData(user)
		queryType := c.DefaultQuery("type", "latest")
		count := c.DefaultQuery("count", "")
		theme := c.DefaultQuery("theme", string(defaultTheme))

		if err != nil {
			log.Printf("Error getting user data for %s: %v", user, err)
			c.String(http.StatusInternalServerError, "Error getting user data: %v", err)
			return
		}
		color, ok := ThemeColorMap[themeFromString(theme)]
		if !ok {
			log.Printf("Color Not Found")
		}
		parsedData, parseErr := parseIt(userData, queryFromString(queryType), countFromString(count), color)
		if parseErr != nil {
			log.Printf("Error parsing data for %s: %v", user, parseErr)
			c.String(http.StatusInternalServerError, "Error parsing data: %v", parseErr)
			return
		}
		c.Data(http.StatusOK, "image/svg+xml", []byte(parsedData))
	})

	return r
}

func main() {
	log.Println("main started")
	r := setupRouter()
	r.Run(":8080")
}
