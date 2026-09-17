package api

import (
	"errors"
	"net/http"

	"github.com/yumestk/TMTClock/internal/store"
)

func registerActivityRoutes(mux *http.ServeMux, s *store.Store) {
	mux.HandleFunc("GET /api/activities", func(w http.ResponseWriter, r *http.Request) {
		acts, err := s.ListActivities(r.Context())
		if err != nil {
			apiErr(w, http.StatusInternalServerError, "list activities: %v", err)
			return
		}
		if acts == nil {
			acts = []store.Activity{}
		}
		writeJSON(w, http.StatusOK, acts)
	})

	mux.HandleFunc("POST /api/activities", func(w http.ResponseWriter, r *http.Request) {
		var body struct {
			Name string `json:"name"`
		}
		if err := readJSON(r, &body); err != nil {
			apiErr(w, http.StatusBadRequest, "invalid body: %v", err)
			return
		}
		if body.Name == "" {
			apiErr(w, http.StatusBadRequest, "name is required")
			return
		}
		a, err := s.CreateActivity(r.Context(), body.Name)
		if err != nil {
			if storeErr(w, err) {
				return
			}
			apiErr(w, http.StatusInternalServerError, "create activity: %v", err)
			return
		}
		writeJSON(w, http.StatusCreated, a)
	})

	mux.HandleFunc("PATCH /api/activities/{id}", func(w http.ResponseWriter, r *http.Request) {
		id, ok := pathID(r)
		if !ok {
			notFound(w, "activity")
			return
		}
		var body struct {
			Name string `json:"name"`
		}
		if err := readJSON(r, &body); err != nil {
			apiErr(w, http.StatusBadRequest, "invalid body: %v", err)
			return
		}
		if body.Name == "" {
			apiErr(w, http.StatusBadRequest, "name is required")
			return
		}
		if err := s.RenameActivity(r.Context(), id, body.Name); err != nil {
			if storeErr(w, err) {
				return
			}
			if errors.Is(err, store.ErrNotFound) {
				notFound(w, "activity")
				return
			}
			apiErr(w, http.StatusInternalServerError, "rename activity: %v", err)
			return
		}
		writeJSON(w, http.StatusOK, store.Activity{ID: id, Name: body.Name, Projects: []store.Project{}})
	})
	mux.HandleFunc("DELETE /api/activities/{id}", func(w http.ResponseWriter, r *http.Request) {
		id, ok := pathID(r)
		if !ok {
			notFound(w, "activity")
			return
		}
		if err := s.DeleteActivity(r.Context(), id); err != nil {
			if storeErr(w, err) {
				return
			}
			if errors.Is(err, store.ErrNotFound) {
				notFound(w, "activity")
				return
			}
			apiErr(w, http.StatusInternalServerError, "delete activity: %v", err)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	})
}
