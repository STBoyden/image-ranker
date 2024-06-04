//go:generate go run github.com/a-h/templ/cmd/templ@latest generate

package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/google/uuid"

	"image-ranker/api"
	"image-ranker/components"
	"image-ranker/consts"
	"image-ranker/static"

	_ "github.com/mattn/go-sqlite3"
)

const databaseFile = "data.db"

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
			log.Println("stopped")
			os.Exit(0)
		}
	}()

	_ = os.Remove(databaseFile)
	db, err := sql.Open("sqlite3", databaseFile)
	if err != nil {
		log.Fatalf("could not open database: %s", err.Error())
	}

	root := components.Root

	mux := http.NewServeMux()
	mux.Handle("/api/", api.Handler(db, tmpStorage))
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
			q := r.URL.Query().Get("requester_id")
			id, err := uuid.Parse(q)
			if err != nil {
				log.Printf("%s, ERR: supplied requester_id query param was not a valid UUIDv4: supplied value was: '%s'", r.RemoteAddr, q)
				log.Printf("%s, >>    redirecting to index with nil requester_id", r.RemoteAddr)
				http.Redirect(w, r, "/", http.StatusFound)
				return
			}

			ctx := context.WithValue(r.Context(), consts.RequesterIDKey, id.String())
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
