package api

import (
	"log"
	"net/http"
	"strings"
)

var handler = http.NewServeMux()

func Handler() http.Handler {
	handler.HandleFunc("/api/", func(w http.ResponseWriter, r *http.Request) {
		path := strings.TrimPrefix(r.URL.Path, "/api")

		log.Printf("serving api response for /api%s", path)

		switch path {
		case "/hello":
			_, _ = w.Write([]byte(`<h1>Hello, bitch</h1>`))
		}
	})

	return handler
}
