package api

import (
	"context"
	"database/sql"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/google/uuid"
	_ "github.com/mattn/go-sqlite3"
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

func showInternalServerError(ctx context.Context, w http.ResponseWriter, format string, v ...any) {
	log.Printf("ERR: "+format, v)
	_ = components.Root(components.Error(ctx, 500, http.StatusText(500))).Render(ctx, w)
}

func GenerateRequesterID(w http.ResponseWriter, r *http.Request) (string, *http.Request, error) {
	id, err := uuid.NewRandom()
	if err != nil {
		showInternalServerError(r.Context(), w, "ERR: somehow could not create UUIDv4: %s", err)
		return "", nil, err
	}

	ctx := context.WithValue(r.Context(), "requesterID", id.String())
	req := r.Clone(ctx)

	return id.String(), req, nil
}

func Handler(database *sql.DB, runtimeImageStoragePath string) http.Handler {
	handler.HandleFunc("/api/", func(w http.ResponseWriter, r *http.Request) {
		err := database.Ping()
		if err != nil {
			showInternalServerError(r.Context(), w, "could not ping database: %s", err)
			return
		}

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
		err = os.Mkdir(imageStoragePath, 0o755)
		if err != nil && !os.IsExist(err) {
			showInternalServerError(r.Context(), w, "somehow could not create path at %s: %s", imageStoragePath, err)
			return
		}
		runtimePaths = append(runtimePaths, imageStoragePath)

		log.Printf("%s, started api Response for /api%s", r.RemoteAddr, path)

		switch path {
		case "/hello":
			_, _ = w.Write([]byte(fmt.Sprintf(`<h1>Hello, there %v</h1>`, requesterID)))
		default:
			log.Printf("%s, served 404 response for path /api%s", r.RemoteAddr, path)
			type Response struct {
				Code    int    `json:"status_code" xml:"StatusCode"`
				Message string `json:"message" xml:"Message"`
			}

			message := Response{Code: 404, Message: http.StatusText(404)}

			w.WriteHeader(404)
			hasSupportedContentType := false
			for _, contentType := range strings.Split(r.Header.Get("Accept"), ",") {
				w.Header().Set("Content-Type", contentType)
				switch contentType {
				case "text/html":
					_ = components.Root(components.Error(r.Context(), message.Code, message.Message)).Render(r.Context(), w)
					hasSupportedContentType = true
					break
				case "text/xml", "application/xml":
					m, _ := xml.Marshal(message)
					_, _ = w.Write(m)
					hasSupportedContentType = true
					break
				}
			}

			if !hasSupportedContentType {
				m, _ := json.Marshal(message)
				w.Header().Set("Content-Type", "application/json")
				_, _ = w.Write(m)
			}
		}
	})

	return handler
}
