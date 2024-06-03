//go:generate go run github.com/a-h/templ/cmd/templ@latest generate

package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"image-ranker/api"
	"image-ranker/components"
	"image-ranker/static"
)

func main() {
	tmpStorage, err := os.MkdirTemp("", "image-ranker-runtime-images")
	if err != nil {
		log.Fatalf("could not create temporary directory: %s", err.Error())
	}

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		for range c {
			log.Println("shutting down...")
			api.Cleanup()
			_ = os.RemoveAll(tmpStorage)
			os.Exit(0)
		}
	}()

	root := components.Root

	mux := http.NewServeMux()
	mux.Handle("/api/", api.Handler(tmpStorage))
	mux.Handle("/static/", static.Handler())
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path

		log.Printf("%s, requested %s", r.RemoteAddr, path)

		if !r.URL.Query().Has("requester_id") {
			var id string
			id, req, err := api.GenerateRequesterID(w, r)
			if err != nil {
				log.Printf("ERR: refer to previous ERR")
				return
			}

			http.Redirect(w, req, path+"?requester_id="+id, http.StatusFound)
			return
		} else {
			ctx := context.WithValue(r.Context(), "requesterID", r.URL.Query().Get("requester_id"))
			r = r.Clone(ctx)
		}

		switch path {
		case "/":
			log.Printf("%s, served /", r.RemoteAddr)
			_ = root(components.Index(r.Context())).Render(r.Context(), w)
		default:
			log.Printf("%s, serving 404 for path %s", r.RemoteAddr, path)
			_ = root(components.Error(r.Context(), 404, fmt.Sprintf("Path \"%s\" not found", path))).Render(r.Context(), w)
			return
		}
	})

	log.Println("Listening on http://localhost:3000")
	_ = http.ListenAndServe(":3000", mux)

	api.Cleanup()
	err = os.RemoveAll(tmpStorage)
	if err != nil {
		log.Fatalf("could not remove temporary directory: %s", err.Error())
	}
}
