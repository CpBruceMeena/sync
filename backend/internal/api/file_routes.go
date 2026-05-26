package api

import (
	"github.com/CpBruceMeena/sync/internal/files"
	"github.com/go-chi/chi/v5"
)

// registerFileRoutes sets up protected file upload route
func registerFileRoutes(r chi.Router, fh *files.Handler) {
	r.Post("/api/files/upload", fh.UploadFile)
}

// registerPublicFileRoutes sets up public file serving route
func registerPublicFileRoutes(r chi.Router, fh *files.Handler) {
	r.Get("/api/files/{filename}", fh.ServeFile)
}
