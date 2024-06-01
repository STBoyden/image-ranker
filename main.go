//go:generate go run github.com/a-h/templ/cmd/templ@latest generate

package main

import (
	"log"
	"net/http"

	"image-ranker/components"

	"github.com/a-h/templ"
)

func main() {
	root := components.Root

	http.Handle("/", templ.Handler(root(components.Index())))
	log.Println("Listening on http://localhost:3000")
	_ = http.ListenAndServe(":3000", nil)
}
