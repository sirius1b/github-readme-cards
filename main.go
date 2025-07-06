package main

import (
	"log"

	. "github.com/sirius1b/github-readme-cards/router"
)

func main() {
	log.Println("main started")
	r := InitRouter(true)
	r.Run(":8080")
}
