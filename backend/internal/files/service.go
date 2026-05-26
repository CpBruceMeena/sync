package files

import (
	"context"
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"

	"github.com/CpBruceMeena/sync/internal/models"
	"github.com/CpBruceMeena/sync/internal/repository"
	"github.com/google/uuid"
)

// Service handles file business logic
type Service struct {
	repos     *repository.Repositories
	uploadDir string
}

// NewService creates a new file service
func NewService(repos *repository.Repositories, uploadDir string) *Service {
	return &Service{repos: repos, uploadDir: uploadDir}
}

// SaveFile saves an uploaded file to disk and creates an attachment record
func (s *Service) SaveFile(ctx context.Context, file multipart.File, header *multipart.FileHeader) (*models.Attachment, error) {
	// Ensure upload directory exists
	if err := os.MkdirAll(s.uploadDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create upload directory: %w", err)
	}

	// Extract extension from original filename
	ext := strings.ToLower(filepath.Ext(header.Filename))

	// Generate unique filename
	filename := uuid.New().String() + ext
	filePath := filepath.Join(s.uploadDir, filename)

	// Create the destination file
	dst, err := os.Create(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to create file: %w", err)
	}
	defer dst.Close()

	// Copy file contents
	if _, err := io.Copy(dst, file); err != nil {
		return nil, fmt.Errorf("failed to write file: %w", err)
	}

	// Detect content type from header (fallback to extension detection)
	contentType := header.Header.Get("Content-Type")
	if contentType == "" {
		contentType = detectContentType(ext)
	}

	// Create attachment record (MessageID will be set later when linked to a message)
	attachment := &models.Attachment{
		ID:       uuid.New(),
		FileUrl:  filename,
		FileType: contentType,
		FileName: header.Filename,
		FileSize: header.Size,
	}

	return attachment, nil
}

// CreateAttachmentRecord saves the attachment record to the database
func (s *Service) CreateAttachmentRecord(ctx context.Context, attachment *models.Attachment) error {
	return s.repos.Attachments.Create(ctx, attachment)
}

// GetFilePath returns the full filesystem path for a given filename
func (s *Service) GetFilePath(filename string) string {
	return filepath.Join(s.uploadDir, filepath.Base(filename))
}

// GetAllowedExtensions returns a list of allowed file extensions
func (s *Service) GetAllowedExtensions() []string {
	return []string{
		".jpg", ".jpeg", ".png", ".gif", ".webp", ".svg",
		".pdf", ".doc", ".docx", ".xls", ".xlsx",
		".txt", ".csv", ".json", ".xml",
		".mp3", ".mp4", ".mov", ".avi",
		".zip", ".tar", ".gz",
	}
}

// MaxFileSize returns the maximum allowed file size in bytes (50MB)
func (s *Service) MaxFileSize() int64 {
	return 50 * 1024 * 1024 // 50MB
}

func detectContentType(ext string) string {
	switch ext {
	case ".jpg", ".jpeg":
		return "image/jpeg"
	case ".png":
		return "image/png"
	case ".gif":
		return "image/gif"
	case ".webp":
		return "image/webp"
	case ".svg":
		return "image/svg+xml"
	case ".pdf":
		return "application/pdf"
	case ".doc":
		return "application/msword"
	case ".docx":
		return "application/vnd.openxmlformats-officedocument.wordprocessingml.document"
	case ".txt":
		return "text/plain"
	case ".csv":
		return "text/csv"
	case ".json":
		return "application/json"
	case ".xml":
		return "application/xml"
	case ".mp3":
		return "audio/mpeg"
	case ".mp4":
		return "video/mp4"
	case ".mov":
		return "video/quicktime"
	case ".zip":
		return "application/zip"
	default:
		return "application/octet-stream"
	}
}
