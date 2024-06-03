//go:generate go run github.com/a-h/templ/cmd/templ@latest generate

package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"image-ranker/api"
	"image-ranker/components"
	"image-ranker/static"
)

func main() {
	tmpStorage, err := os.MkdirTemp("", "image-ranker-runtime-images")
	if err != nil {
		log.Fatalf("could not create temporary directory: %s", err.Error())
	}

	root := components.Root

	mux := http.NewServeMux()
	mux.Handle("/api/", api.Handler())
	mux.Handle("/static/", static.Handler())
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path
		context := r.Context()

		log.Printf("requested %s", path)

		switch path {
		case "/":
			log.Println("serving /")
			_ = root(components.Index()).Render(context, w)
		default:
			log.Printf("serving 404 for path %s", path)
			_ = root(components.Error(404, fmt.Sprintf("Path \"%s\" not found", path))).Render(context, w)
		}
	})

	log.Println("Listening on http://localhost:3000")
	_ = http.ListenAndServe(":3000", mux)

	err = os.RemoveAll(tmpStorage)
	if err != nil {
		log.Fatalf("could not remove temporary directory: %s", err.Error())
	}
}
