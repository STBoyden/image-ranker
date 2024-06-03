package api

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/google/uuid"
	"image-ranker/components"
)

var (
	handler      = http.NewServeMux()
	runtimePaths []string
)

func Cleanup() {
	for _, path := range runtimePaths {
		err := os.RemoveAll(path)
		if err != nil {
			log.Printf("ERR: could not remove path %s: %s", path, err)
		}
	}
}

func GenerateRequesterID(w http.ResponseWriter, r *http.Request) (string, *http.Request, error) {
	id, err := uuid.NewRandom()
	if err != nil {
		log.Printf("ERR: somehow could not create UUIDv4: %s", err)
		_ = components.Root(components.Error(r.Context(), 500, "an internal error occurred, please notify the server administrator")).Render(r.Context(), w)
		return "", nil, err
	}

	ctx := context.WithValue(r.Context(), "requesterID", id.String())
	req := r.Clone(ctx)

	return id.String(), req, nil
}

func Handler(runtimeImageStoragePath string) http.Handler {
	handler.HandleFunc("/api/", func(w http.ResponseWriter, r *http.Request) {
		root := components.Root
		path := strings.TrimPrefix(r.URL.Path, "/api")

		var requesterID string
		if r.URL.Query().Get("requester_id") != "" {
			requesterID = r.URL.Query().Get("requester_id")
		} else if r.Context().Value("requesterID") != nil {
			requesterID = r.Context().Value("requesterID").(string)
		} else {
			requesterID, r, _ = GenerateRequesterID(w, r)
		}

		imageStoragePath := filepath.Join(runtimeImageStoragePath, requesterID)
		err := os.Mkdir(imageStoragePath, 0o755)
		if err != nil && !os.IsExist(err) {
			log.Printf("ERR: somehow could not create path at %s: %s", imageStoragePath, err)
			_ = root(components.Error(r.Context(), 500, "an internal error occurred, please notify the server administrator")).Render(r.Context(), w)
			return
		}
		runtimePaths = append(runtimePaths, imageStoragePath)

		log.Printf("%s, serving api response for /api%s", r.RemoteAddr, path)

		switch path {
		case "/hello":
			_, _ = w.Write([]byte(fmt.Sprintf(`<h1>Hello, there %v</h1>`, requesterID)))
		}
	})

	return handler
}
