package files

import "github.com/google/uuid"

// Handler handles file upload and serving HTTP requests
type Handler struct {
	service   *Service
	uploadDir string
}

// NewHandler creates a new files HTTP handler
func NewHandler(svc *Service, uploadDir string) *Handler {
	return &Handler{service: svc, uploadDir: uploadDir}
}

// UploadResponse represents the response after a file upload
type UploadResponse struct {
	Attachment AttachmentResponse `json:"attachment"`
	FileURL    string             `json:"file_url"`
	FileName   string             `json:"file_name"`
	FileType   string             `json:"file_type"`
	FileSize   int64              `json:"file_size"`
}

// AttachmentResponse represents an attachment in API responses
type AttachmentResponse struct {
	ID       uuid.UUID `json:"id"`
	FileURL  string    `json:"file_url"`
	FileType string    `json:"file_type"`
	FileName string    `json:"file_name"`
	FileSize int64     `json:"file_size"`
}
