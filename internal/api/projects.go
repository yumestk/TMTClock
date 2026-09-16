package api

import (
	"errors"
	"net/http"

	"github.com/yumestk/TMTClock/internal/store"
)

func registerProjectRoutes(mux *http.ServeMux, s *store.Store) {
	mux.HandleFunc("POST /api/projects", func(w http.ResponseWriter, r *http.Request) {
		var body struct {
			ActivityID      int64  `json:"activity_id"`
			Name            string `json:"name"`
			ExpectedMinutes int    `json:"expected_minutes"`
		}
		if err := readJSON(r, &body); err != nil {
			apiErr(w, http.StatusBadRequest, "invalid body: %v", err)
			return
		}
		if body.Name == "" {
			apiErr(w, http.StatusBadRequest, "name is required")
			return
		}
		if body.ExpectedMinutes < 0 {
			apiErr(w, http.StatusBadRequest, "expected_minutes must be >= 0")
			return
		}
		p, err := s.CreateProject(r.Context(), body.ActivityID, body.Name, body.ExpectedMinutes)
		if err != nil {
			if storeErr(w, err) {
				return
			}
			if errors.Is(err, store.ErrNotFound) {
				notFound(w, "activity")
				return
			}
			apiErr(w, http.StatusInternalServerError, "create project: %v", err)
			return
		}
		writeJSON(w, http.StatusCreated, p)
	})

	mux.HandleFunc("PATCH /api/projects/{id}", func(w http.ResponseWriter, r *http.Request) {
		id, ok := pathID(r)
		if !ok {
			notFound(w, "project")
			return
		}
		var body struct {
			Name            *string `json:"name"`
			ExpectedMinutes *int    `json:"expected_minutes"`
		}
		if err := readJSON(r, &body); err != nil {
			apiErr(w, http.StatusBadRequest, "invalid body: %v", err)
			return
		}
		if err := s.UpdateProject(r.Context(), id, body.Name, body.ExpectedMinutes); err != nil {
			if storeErr(w, err) {
				return
			}
			if errors.Is(err, store.ErrNotFound) {
				notFound(w, "project")
				return
			}
			apiErr(w, http.StatusInternalServerError, "update project: %v", err)
			return
		}
		p, err := s.GetProject(r.Context(), id)
		if err != nil {
			apiErr(w, http.StatusInternalServerError, "reload project: %v", err)
			return
		}
		writeJSON(w, http.StatusOK, p)
	})

	mux.HandleFunc("DELETE /api/projects/{id}", func(w http.ResponseWriter, r *http.Request) {
		id, ok := pathID(r)
		if !ok {
			notFound(w, "project")
			return
		}
		if err := s.DeleteProject(r.Context(), id); err != nil {
			if storeErr(w, err) {
				return
			}
			if errors.Is(err, store.ErrNotFound) {
				notFound(w, "project")
				return
			}
			apiErr(w, http.StatusInternalServerError, "delete project: %v", err)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	})
}
