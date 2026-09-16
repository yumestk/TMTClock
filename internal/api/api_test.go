package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"
	"time"

	"github.com/yumestk/TMTClock/internal/store"
)

func newTestServer(t *testing.T) *httptest.Server {
	t.Helper()
	s, err := store.Open(filepath.Join(t.TempDir(), "api-test.sqlite"))
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	t.Cleanup(func() { s.Close() })
	srv := httptest.NewServer(New(s))
	t.Cleanup(srv.Close)
	return srv
}

func do(t *testing.T, srv *httptest.Server, method, path string, body any) *http.Response {
	t.Helper()
	var rd *bytes.Reader
	if body != nil {
		bs, err := json.Marshal(body)
		if err != nil {
			t.Fatalf("marshal: %v", err)
		}
		rd = bytes.NewReader(bs)
	} else {
		rd = bytes.NewReader(nil)
	}
	req, err := http.NewRequest(method, srv.URL+path, rd)
	if err != nil {
		t.Fatalf("new request: %v", err)
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	res, err := srv.Client().Do(req)
	if err != nil {
		t.Fatalf("%s %s: %v", method, path, err)
	}
	return res
}

func decode(t *testing.T, res *http.Response, v any) {
	t.Helper()
	defer res.Body.Close()
	if err := json.NewDecoder(res.Body).Decode(v); err != nil {
		t.Fatalf("decode response: %v", err)
	}
}

func TestFullFlow(t *testing.T) {
	srv := newTestServer(t)

	// Create activity + project.
	var act store.Activity
	decode(t, do(t, srv, "POST", "/api/activities", map[string]string{"name": "学习"}), &act)
	var proj store.Project
	decode(t, do(t, srv, "POST", "/api/projects", map[string]any{
		"activity_id": act.ID, "name": "英语", "expected_minutes": 25,
	}), &proj)

	// Running: none.
	if res := do(t, srv, "GET", "/api/running", nil); res.StatusCode != http.StatusNoContent {
		t.Fatalf("running with none: status %d", res.StatusCode)
	}

	// Start.
	var started store.Session
	res := do(t, srv, "POST", "/api/sessions/start", map[string]any{
		"project_id": proj.ID, "note": "计划", "expected_minutes": 25,
	})
	if res.StatusCode != http.StatusCreated {
		t.Fatalf("start: status %d", res.StatusCode)
	}
	decode(t, res, &started)
	if started.EndAt != nil || started.ActivityName != "学习" || started.ProjectName != "英语" {
		t.Fatalf("started session wrong: %+v", started)
	}

	// Second start -> 409.
	if res := do(t, srv, "POST", "/api/sessions/start", map[string]any{"project_id": proj.ID}); res.StatusCode != http.StatusConflict {
		t.Fatalf("second start: status %d", res.StatusCode)
	}

	// Running: yes, same id.
	var running store.Session
	decode(t, do(t, srv, "GET", "/api/running", nil), &running)
	if running.ID != started.ID {
		t.Fatalf("running id %d, want %d", running.ID, started.ID)
	}

	// Today shows the running row.
	var today struct {
		Date     string          `json:"date"`
		Sessions []store.Session `json:"sessions"`
	}
	decode(t, do(t, srv, "GET", "/api/today", nil), &today)
	if today.Date != time.Now().Format("2006-01-02") {
		t.Fatalf("today date %q", today.Date)
	}
	if len(today.Sessions) != 1 || today.Sessions[0].EndAt != nil {
		t.Fatalf("today should show the running session: %+v", today.Sessions)
	}

	// Stop with note override.
	var stopped store.Session
	res = do(t, srv, "POST", "/api/sessions/stop", map[string]any{"note": "实际做的事"})
	if res.StatusCode != http.StatusOK {
		t.Fatalf("stop: status %d", res.StatusCode)
	}
	decode(t, res, &stopped)
	if stopped.EndAt == nil || stopped.Note != "实际做的事" {
		t.Fatalf("stopped session wrong: %+v", stopped)
	}

	// Stop again -> 409.
	if res := do(t, srv, "POST", "/api/sessions/stop", nil); res.StatusCode != http.StatusConflict {
		t.Fatalf("second stop: status %d", res.StatusCode)
	}

	// Stop with empty body (no JSON) works.
	// (already covered by 409 above; malformed JSON next)
	if res := do(t, srv, "POST", "/api/sessions/stop", nil); res.StatusCode != http.StatusConflict {
		t.Fatalf("stop empty body: status %d", res.StatusCode)
	}

	// Today shows the completed row.
	decode(t, do(t, srv, "GET", "/api/today", nil), &today)
	if len(today.Sessions) != 1 || today.Sessions[0].EndAt == nil {
		t.Fatalf("today should show the completed session: %+v", today.Sessions)
	}

	// Start with missing project -> 404.
	if res := do(t, srv, "POST", "/api/sessions/start", map[string]any{"project_id": 999}); res.StatusCode != http.StatusNotFound {
		t.Fatalf("start missing project: status %d", res.StatusCode)
	}
}

func TestBadRequests(t *testing.T) {
	srv := newTestServer(t)

	if res := do(t, srv, "POST", "/api/activities", map[string]string{"name": ""}); res.StatusCode != http.StatusBadRequest {
		t.Fatalf("empty name: status %d", res.StatusCode)
	}

	req, _ := http.NewRequest("POST", srv.URL+"/api/activities", bytes.NewReader([]byte(`{"name":`)))
	req.Header.Set("Content-Type", "application/json")
	res, err := srv.Client().Do(req)
	if err != nil {
		t.Fatalf("malformed request: %v", err)
	}
	if res.StatusCode != http.StatusBadRequest {
		t.Fatalf("malformed json: status %d", res.StatusCode)
	}

	if res := do(t, srv, "POST", "/api/activities", map[string]any{"name": "x", "bogus": 1}); res.StatusCode != http.StatusBadRequest {
		t.Fatalf("unknown field: status %d", res.StatusCode)
	}

	if res := do(t, srv, "PATCH", fmt.Sprintf("/api/activities/%s", "abc"), nil); res.StatusCode != http.StatusNotFound {
		t.Fatalf("non-numeric id: status %d", res.StatusCode)
	}
}
