package util

import (
	"bytes"
	"fmt"
	"log"
	"strconv"
	"strings"

	. "github.com/sirius1b/github-readme-cards/internal"
	"golang.org/x/net/html"
)

func ParseIt(rss RSS2JSONResponse, queryType QueryType, count int, color ThemeColors) (string, error) {
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

func HomePage() string {
	return `
	<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8" />
  <meta name="viewport" content="width=device-width, initial-scale=1" />
  <title>GitHub Readme Medium Cards</title>
  <link rel="stylesheet" href="https://cdn.jsdelivr.net/npm/@picocss/pico@1/css/pico.min.css" />
</head>
<body>

  <!-- Navigation -->
  <nav class="container-fluid">
    <ul>
      <li><strong>GitHub Readme Medium Cards</strong></li>
    </ul>
    <ul>
      <li><a href="https://github.com/sirius1b/github-readme-cards" target="_blank">GitHub</a></li>
      <li><a href="#usage">Usage</a></li>
      <li><a href="#themes" role="button">Themes</a></li>
    </ul>
  </nav>

  <!-- Main Content -->
  <main class="container">
    <div class="grid">
      <section>
        <hgroup>
          <h2>✨ Embed Your Medium Blog in Your GitHub Profile</h2>
          <h3>Auto-updating, themed blog cards for your GitHub README</h3>
        </hgroup>
        <p>
          This tool lets you showcase your latest Medium articles directly in your GitHub profile using a styled, minimal SVG card.
          Just copy the usage URL and customize it with your Medium username and preferred theme.
        </p>

        <h3 id="usage">🔧 Usage</h3>
        <p>Use this endpoint:</p>
        <pre><code>https://github-readme-cards.vercel.app/medium/user/&lt;username&gt;?count=&lt;count&gt;&theme=&lt;theme&gt;</code></pre>

        <h4>Parameters:</h4>
        <ul>
          <li><code>username</code>: Your Medium username (e.g. <code>lav.nya.verma</code>)</li>
          <li><code>count</code>: Number of posts to display (1-10)</li>
          <li><code>theme</code>: Visual theme for the card</li>
        </ul>

        <h4>Example:</h4>
        <pre><code>https://github-readme-cards.vercel.app/medium/user/lav.nya.verma?count=5&theme=tokyonight</code></pre>

        <h3 id="themes">🎨 Available Themes</h3>
        <p>Choose from a range of popular developer themes:</p>

        <div class="grid">
          <div>
            <strong>🌞 Light Themes</strong>
            <ul>
              <li><code>solarizedlight</code></li>
              <li><code>githublight</code></li>
              <li><code>rosepinedawn</code></li>
              <li><code>quietlight</code></li>
            </ul>
          </div>
          <div>
            <strong>🌙 Dark Themes</strong>
            <ul>
              <li><code>dracula</code></li>
              <li><code>gruvbox</code></li>
              <li><code>nord</code></li>
              <li><code>onedark</code></li>
              <li><code>monokai</code></li>
              <li><code>tokyonight</code></li>
            </ul>
          </div>
        </div>

        <h3>🖼 Preview</h3>
        <p>Paste the URL in your browser or embed in your README:</p>
        <pre><code>![Medium Blog](https://github-readme-cards.vercel.app/medium/user/lav.nya.verma?count=5&theme=onedark)</code></pre>
        <img src="https://github-readme-cards.vercel.app/medium/user/lav.nya.verma?count=3&theme=onedark" alt="Preview" style="max-width: 100%; border-radius: 12px; box-shadow: 0 1px 4px rgba(0,0,0,0.1);" />

      </section>
    </div>
  </main>

  <!-- Feature Request Section -->
  <section aria-label="Feature Request">
    <div class="container">
      <article>
        <hgroup>
          <h2>💡 Have a Feature Request or Bug Report?</h2>
          <h3>Open an issue on the GitHub repo</h3>
        </hgroup>
        <p>
          We welcome contributions, improvements, and feedback! If you have ideas for new features or spot any bugs, click the button below to submit an issue.
        </p>
        <a href="https://github.com/sirius1b/github-readme-cards/issues" class="contrast" role="button" target="_blank">Submit an Issue</a>
      </article>
    </div>
  </section>

  <!-- Footer -->
  <footer class="container">
    <small>
      <a href="https://github.com/sirius1b/github-readme-cards">GitHub Repo</a> • 
      <a href="https://vercel.com">Powered by Vercel</a>
    </small>
  </footer>

</body>
</html>

	`
}
