package api

import (
	"github.com/CpBruceMeena/sync/internal/discovery"
	"github.com/go-chi/chi/v5"
)

func registerDiscoveryRoutes(r chi.Router, dh *discovery.Handler) {
	r.Get("/api/discovery/users", dh.SearchUsers)
	r.Get("/api/discovery/groups", dh.ListPublicGroups)
	r.Get("/api/discovery/groups/search", dh.SearchGroups)
	r.Get("/api/discovery/groups/{id}", dh.GetGroupDetails)
}
