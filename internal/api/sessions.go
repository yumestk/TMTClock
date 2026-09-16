package api

import (
	"net/http"
	"time"

	"github.com/yumestk/TMTClock/internal/store"
)

func registerSessionRoutes(mux *http.ServeMux, s *store.Store) {
	mux.HandleFunc("GET /api/running", func(w http.ResponseWriter, r *http.Request) {
		sess, running, err := s.RunningSession(r.Context())
		if err != nil {
			apiErr(w, http.StatusInternalServerError, "get running: %v", err)
			return
		}
		if !running {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		writeJSON(w, http.StatusOK, sess)
	})

	mux.HandleFunc("POST /api/sessions/start", func(w http.ResponseWriter, r *http.Request) {
		var body struct {
			ProjectID       int64  `json:"project_id"`
			Note            string `json:"note"`
			ExpectedMinutes int    `json:"expected_minutes"`
		}
		if err := readJSON(r, &body); err != nil {
			apiErr(w, http.StatusBadRequest, "invalid body: %v", err)
			return
		}
		if body.ExpectedMinutes < 0 {
			apiErr(w, http.StatusBadRequest, "expected_minutes must be >= 0")
			return
		}
		sess, err := s.StartSession(r.Context(), body.ProjectID, body.Note, body.ExpectedMinutes)
		if err != nil {
			if storeErr(w, err) {
				return
			}
			if isNotFound(err) {
				notFound(w, "project")
				return
			}
			apiErr(w, http.StatusInternalServerError, "start session: %v", err)
			return
		}
		writeJSON(w, http.StatusCreated, sess)
	})

	mux.HandleFunc("POST /api/sessions/stop", func(w http.ResponseWriter, r *http.Request) {
		var body struct {
			Note *string `json:"note"`
		}
		// An empty body is valid for stop.
		if r.ContentLength != 0 {
			if err := readJSON(r, &body); err != nil {
				apiErr(w, http.StatusBadRequest, "invalid body: %v", err)
				return
			}
		}
		sess, err := s.StopSession(r.Context(), body.Note)
		if err != nil {
			if storeErr(w, err) {
				return
			}
			apiErr(w, http.StatusInternalServerError, "stop session: %v", err)
			return
		}
		writeJSON(w, http.StatusOK, sess)
	})

	mux.HandleFunc("GET /api/today", func(w http.ResponseWriter, r *http.Request) {
		from, to := dayBounds(time.Now())
		sessions, err := s.ListDaySessions(r.Context(), from, to)
		if err != nil {
			apiErr(w, http.StatusInternalServerError, "list day sessions: %v", err)
			return
		}
		if sessions == nil {
			sessions = []store.Session{}
		}
		writeJSON(w, http.StatusOK, map[string]any{
			"date":     time.Now().Format("2006-01-02"),
			"sessions": sessions,
		})
	})
}

// dayBounds returns [from, to) UTC instants covering the local calendar day
// of now. time.Date normalizes day+1, so DST-shortened days are handled.
func dayBounds(now time.Time) (time.Time, time.Time) {
	loc := now.Location()
	y, m, d := now.Date()
	from := time.Date(y, m, d, 0, 0, 0, 0, loc)
	to := time.Date(y, m, d+1, 0, 0, 0, 0, loc)
	return from.UTC(), to.UTC()
}
