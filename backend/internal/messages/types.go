package messages

import "github.com/CpBruceMeena/sync/internal/database"

// Handler handles message HTTP requests
type Handler struct {
	queries database.Querier
}
