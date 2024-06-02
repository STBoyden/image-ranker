//go:generate go run github.com/a-h/templ/cmd/templ@latest generate

package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"image-ranker/components"
)

func main() {
	tmpStorage, err := os.MkdirTemp("", "image-ranker-runtime-images")
	if err != nil {
		panic("could not create temporary directory: " + err.Error())
	}

	root := components.Root

	mux := http.NewServeMux()
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
		panic("could not remove temporary directory: " + err.Error())
	}
}
