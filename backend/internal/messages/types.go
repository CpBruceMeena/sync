package messages

import (
	"github.com/CpBruceMeena/sync/internal/repository"
)

// Handler handles message HTTP requests
type Handler struct {
	repos *repository.Repositories
}
