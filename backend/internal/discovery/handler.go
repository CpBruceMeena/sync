package discovery

import (
	"log"
	"net/http"
	"strconv"

	"github.com/CpBruceMeena/sync/internal/httputil"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

// SearchUsers handles GET /api/discovery/users?q=term&limit=20
// @Summary Search users
// @Description Search for users by username, display name, or email
// @Tags discovery
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param q query string true "Search query"
// @Param limit query int false "Max results (default 20, max 50)"
// @Success 200 {array} UserResult
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/discovery/users [get]
func (h *Handler) SearchUsers(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query().Get("q")
	if q == "" {
		httputil.RespondError(w, http.StatusBadRequest, "Query parameter 'q' is required")
		return
	}

	limit := parseLimit(r.URL.Query().Get("limit"))

	users, err := h.service.SearchUsers(r.Context(), q, limit)
	if err != nil {
		log.Printf("Error searching users: %v", err)
		httputil.RespondError(w, http.StatusInternalServerError, "Failed to search users")
		return
	}

	httputil.RespondJSON(w, http.StatusOK, users)
}

// ListPublicGroups handles GET /api/discovery/groups?limit=20&offset=0
// @Summary List public groups
// @Description List all public groups with pagination
// @Tags discovery
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param limit query int false "Results per page (default 20, max 50)"
// @Param offset query int false "Offset for pagination (default 0)"
// @Success 200 {array} GroupDetailResponse
// @Failure 500 {object} map[string]string
// @Router /api/discovery/groups [get]
func (h *Handler) ListPublicGroups(w http.ResponseWriter, r *http.Request) {
	limit := parseLimit(r.URL.Query().Get("limit"))
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))

	groups, err := h.service.ListPublicGroups(r.Context(), limit, offset)
	if err != nil {
		log.Printf("Error listing public groups: %v", err)
		httputil.RespondError(w, http.StatusInternalServerError, "Failed to list public groups")
		return
	}

	httputil.RespondJSON(w, http.StatusOK, groups)
}

// SearchGroups handles GET /api/discovery/groups?q=term&limit=20
// @Summary Search public groups
// @Description Search for public groups by name
// @Tags discovery
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param q query string true "Search query"
// @Param limit query int false "Max results (default 20, max 50)"
// @Success 200 {array} GroupDetailResponse
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/discovery/groups/search [get]
func (h *Handler) SearchGroups(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query().Get("q")
	if q == "" {
		httputil.RespondError(w, http.StatusBadRequest, "Query parameter 'q' is required")
		return
	}

	limit := parseLimit(r.URL.Query().Get("limit"))

	groups, err := h.service.SearchPublicGroups(r.Context(), q, limit)
	if err != nil {
		log.Printf("Error searching groups: %v", err)
		httputil.RespondError(w, http.StatusInternalServerError, "Failed to search groups")
		return
	}

	httputil.RespondJSON(w, http.StatusOK, groups)
}

// GetGroupDetails handles GET /api/discovery/groups/{id}
// @Summary Get group details
// @Description Get detailed information about a public group
// @Tags discovery
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Group conversation ID"
// @Success 200 {object} GroupDetailResponse
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Router /api/discovery/groups/{id} [get]
func (h *Handler) GetGroupDetails(w http.ResponseWriter, r *http.Request) {
	convIDStr := chi.URLParam(r, "id")
	convID, err := uuid.Parse(convIDStr)
	if err != nil {
		httputil.RespondError(w, http.StatusBadRequest, "Invalid group ID")
		return
	}

	group, err := h.service.GetGroupDetails(r.Context(), convID)
	if err != nil {
		httputil.RespondError(w, http.StatusNotFound, "Group not found")
		return
	}
	if group == nil {
		httputil.RespondError(w, http.StatusNotFound, "Group not found or is not public")
		return
	}

	httputil.RespondJSON(w, http.StatusOK, group)
}

func parseLimit(s string) int {
	if s == "" {
		return 20
	}
	limit, err := strconv.Atoi(s)
	if err != nil || limit <= 0 {
		return 20
	}
	if limit > 50 {
		return 50
	}
	return limit
}
