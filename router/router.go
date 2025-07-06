package router

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"

	. "github.com/sirius1b/github-readme-cards/client"
	. "github.com/sirius1b/github-readme-cards/internal"
	. "github.com/sirius1b/github-readme-cards/util"
)

var db = make(map[string]FeedResponse)

func InitRouter(release bool) *gin.Engine {
	log.Println("setupRouter called")
	r := gin.Default()

	if release {
		gin.SetMode(gin.ReleaseMode)
		log.Println("Running in release mode")
	}

	r.GET("/", func(c *gin.Context) {
		log.Println("GET / called")
		c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(HomePage()))
	})

	r.GET("/medium/user/:name", func(c *gin.Context) {
		user := c.Params.ByName("name")
		log.Printf("GET /user/%s called", user)
		userData, err := GetUserData(user, db)

		if err != nil {
			log.Printf("Error getting user data for %s: %v", user, err)
			c.String(http.StatusInternalServerError, "Error getting user data: %v", err)
			return
		}

		queryType := c.DefaultQuery("type", "latest")
		count := c.DefaultQuery("count", "")
		theme := c.DefaultQuery("theme", string(DefaultTheme))

		color, ok := ThemeColorMap[ThemeFromString(theme)]
		if !ok {
			log.Printf("Color Not Found")
		}
		parsedData, parseErr := ParseIt(userData, QueryFromString(queryType), CountFromString(count), color)
		if parseErr != nil {
			log.Printf("Error parsing data for %s: %v", user, parseErr)
			c.String(http.StatusInternalServerError, "Error parsing data: %v", parseErr)
			return
		}
		c.Data(http.StatusOK, "image/svg+xml", []byte(parsedData))
	})

	r.NoRoute(func(ctx *gin.Context) {
		log.Println("No route matched, redirecting to home page")
		ctx.Redirect(http.StatusMovedPermanently, "/")
	})

	return r
}
