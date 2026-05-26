package api

import (
	"github.com/CpBruceMeena/sync/internal/files"
	"github.com/go-chi/chi/v5"
)

// registerFileRoutes sets up file upload and serving routes
func registerFileRoutes(r chi.Router, fh *files.Handler) {
	r.Post("/api/files/upload", fh.UploadFile)
	r.Get("/api/files/{filename}", fh.ServeFile)
}
