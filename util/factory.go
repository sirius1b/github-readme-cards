package util

import (
	"log"
	"strconv"
	"strings"

	. "github.com/sirius1b/github-readme-cards/internal"
)

func QueryFromString(query string) QueryType {
	log.Printf("queryFromString called with query: %s", query)
	switch query {
	case "latest":
		return Latest
	default:
		return DefaultType // Default to Latest if unknown
	}
}

func CountFromString(countStr string) int {
	log.Printf("countFromString called with countStr: %s", countStr)
	if countStr == "" {
		return DefaultMaxCount
	}
	count, err := strconv.Atoi(countStr)
	if err != nil || count <= 0 {
		log.Printf("Invalid count value: %s, defaulting to %d", countStr, DefaultMaxCount)
		return DefaultMaxCount
	}
	return count
}

func ThemeFromString(themeStr string) ColorTheme {
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
		return DefaultTheme // Default to GithubLight if unknown
	}
}
