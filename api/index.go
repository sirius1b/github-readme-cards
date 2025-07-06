package handler

import (
	"net/http"

	. "github.com/sirius1b/github-readme-cards/router"
)

func Handler(w http.ResponseWriter, r *http.Request) {
	router := InitRouter(true)
	router.ServeHTTP(w, r)
}
