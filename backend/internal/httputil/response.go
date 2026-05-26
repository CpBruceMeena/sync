package httputil

import (
	"encoding/json"
	"net/http"
)

// RespondJSON sends a JSON response with the given status code and data
func RespondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

// RespondError sends a JSON error response with the given status code and message
func RespondError(w http.ResponseWriter, status int, message string) {
	RespondJSON(w, status, map[string]string{"error": message})
}
