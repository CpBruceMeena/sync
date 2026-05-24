package conversations

import "github.com/CpBruceMeena/sync/internal/database"

// Handler handles conversation HTTP requests
type Handler struct {
	queries database.Querier
}
