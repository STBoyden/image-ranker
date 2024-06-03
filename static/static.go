package static

import (
	"fmt"
	"log"
	"net/http"
	"strings"
)

var (
	handler         = http.NewServeMux()
	disallowedPaths = []string{"static.go"}
)

func Handler() http.Handler {
	handler.HandleFunc("/static/", func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path

		for _, p := range disallowedPaths {
			if strings.Contains(path, p) {
				log.Printf("path %s is not allowed. redirecting to /%s...", path, p)
				http.Redirect(w, r, fmt.Sprintf("/%s", p), http.StatusSeeOther)
				return
			}
		}

		log.Printf("%s, served %s", r.RemoteAddr, path)
		http.StripPrefix("/static/", http.FileServer(http.Dir(".")))
	})

	return handler
}
