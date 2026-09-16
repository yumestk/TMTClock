// Package api exposes the store over a small JSON HTTP interface.
package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/yumestk/TMTClock/internal/store"
)

// New builds the ServeMux with all routes registered.
func New(s *store.Store) *http.ServeMux {
	mux := http.NewServeMux()
	registerActivityRoutes(mux, s)
	registerProjectRoutes(mux, s)
	return mux
}

func writeJSON(w http.ResponseWriter, code int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	if v != nil {
		_ = json.NewEncoder(w).Encode(v)
	}
}

func readJSON(r *http.Request, v any) error {
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	return dec.Decode(v)
}

func apiErr(w http.ResponseWriter, code int, format string, args ...any) {
	writeJSON(w, code, map[string]string{"error": fmt.Sprintf(format, args...)})
}

// storeErr maps a store error to an HTTP status, reporting whether it was mapped.
func storeErr(w http.ResponseWriter, err error) bool {
	switch {
	case errors.Is(err, store.ErrDuplicate):
		apiErr(w, http.StatusConflict, "%s", err)
	case errors.Is(err, store.ErrHasSessions):
		apiErr(w, http.StatusConflict, "%s", err)
	case errors.Is(err, store.ErrAlreadyRunning), errors.Is(err, store.ErrNotRunning):
		apiErr(w, http.StatusConflict, "%s", err)
	default:
		return false
	}
	return true
}

func notFound(w http.ResponseWriter, what string) {
	apiErr(w, http.StatusNotFound, "%s not found", what)
}

func pathID(r *http.Request) (int64, bool) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil || id <= 0 {
		return 0, false
	}
	return id, true
}
