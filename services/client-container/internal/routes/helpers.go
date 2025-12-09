package routes

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi"
)

// writeJSON writes a JSON response
func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		http.Error(w, "failed to encode response", http.StatusInternalServerError)
	}
}

// writeError writes an error response
func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{
		"error": message,
	})
}

// getIDParam extracts an integer ID parameter from the URL
func getIDParam(r *http.Request, param string) (int, error) {
	idStr := chi.URLParam(r, param)
	id, err := strconv.Atoi(idStr)
	if err != nil {
		return 0, err
	}
	return id, nil
}
