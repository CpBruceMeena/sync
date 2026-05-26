package files

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/CpBruceMeena/sync/internal/httputil"
	"github.com/go-chi/chi/v5"
)

const maxUploadSize = 50 << 20 // 50MB

// UploadFile handles file upload requests
// @Summary Upload file
// @Description Upload a file (image, document, etc.) and return the file metadata
// @Tags files
// @Accept multipart/form-data
// @Produce json
// @Security BearerAuth
// @Param file formData file true "File to upload"
// @Success 201 {object} UploadResponse
// @Failure 400 {object} map[string]string
// @Failure 413 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/files/upload [post]
func (h *Handler) UploadFile(w http.ResponseWriter, r *http.Request) {
	// Limit request body size
	r.Body = http.MaxBytesReader(w, r.Body, maxUploadSize)

	if err := r.ParseMultipartForm(maxUploadSize); err != nil {
		httputil.RespondError(w, http.StatusBadRequest, "File too large or invalid form data (max 50MB)")
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		httputil.RespondError(w, http.StatusBadRequest, "No file provided in 'file' field")
		return
	}
	defer file.Close()

	// Validate file extension
	ext := strings.ToLower(filepath.Ext(header.Filename))
	allowed := false
	for _, allowedExt := range h.service.GetAllowedExtensions() {
		if ext == allowedExt {
			allowed = true
			break
		}
	}
	if !allowed {
		httputil.RespondError(w, http.StatusBadRequest, fmt.Sprintf("File type '%s' is not allowed", ext))
		return
	}

	// Save file and create attachment record
	attachment, err := h.service.SaveFile(r.Context(), file, header)
	if err != nil {
		log.Printf("Error saving file: %v", err)
		httputil.RespondError(w, http.StatusInternalServerError, "Failed to save file")
		return
	}

	// Save attachment record to database
	if err := h.service.CreateAttachmentRecord(r.Context(), attachment); err != nil {
		log.Printf("Error saving attachment record: %v", err)
		// Try to clean up the file
		os.Remove(h.service.GetFilePath(attachment.FileUrl))
		httputil.RespondError(w, http.StatusInternalServerError, "Failed to save attachment record")
		return
	}

	resp := UploadResponse{
		Attachment: AttachmentResponse{
			ID:       attachment.ID,
			FileURL:  attachment.FileUrl,
			FileType: attachment.FileType,
			FileName: attachment.FileName,
			FileSize: attachment.FileSize,
		},
		FileURL:  attachment.FileUrl,
		FileName: attachment.FileName,
		FileType: attachment.FileType,
		FileSize: attachment.FileSize,
	}

	httputil.RespondJSON(w, http.StatusCreated, resp)
}

// ServeFile serves uploaded files
// @Summary Get file
// @Description Retrieve an uploaded file by filename
// @Tags files
// @Produce application/octet-stream
// @Param filename path string true "Filename"
// @Success 200 {file} binary
// @Failure 404 {object} map[string]string
// @Router /api/files/{filename} [get]
func (h *Handler) ServeFile(w http.ResponseWriter, r *http.Request) {
	filename := chi.URLParam(r, "filename")

	if filename == "" || strings.Contains(filename, "..") {
		httputil.RespondError(w, http.StatusBadRequest, "Invalid filename")
		return
	}

	filePath := h.service.GetFilePath(filename)

	// Check if file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		httputil.RespondError(w, http.StatusNotFound, "File not found")
		return
	}

	// Open and serve the file
	file, err := os.Open(filePath)
	if err != nil {
		log.Printf("Error opening file: %v", err)
		httputil.RespondError(w, http.StatusInternalServerError, "Failed to open file")
		return
	}
	defer file.Close()

	// Determine content type from filename
	ext := strings.ToLower(filepath.Ext(filename))
	contentType := detectContentType(ext)

	w.Header().Set("Content-Type", contentType)
	w.Header().Set("Content-Disposition", fmt.Sprintf(`inline; filename="%s"`, filepath.Base(filename)))

	io.Copy(w, file)
}

// NewUploadHandler creates an http.Handler that limits request size for file uploads
func NewUploadHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/api/files/upload") {
			r.Body = http.MaxBytesReader(w, r.Body, maxUploadSize)
		}
		next.ServeHTTP(w, r)
	})
}
