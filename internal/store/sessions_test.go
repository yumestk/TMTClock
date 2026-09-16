package store

import (
	"context"
	"errors"
	"testing"
	"time"
)

func setupSessionFixture(t *testing.T) (*Store, Project) {
	t.Helper()
	s := newTestStore(t)
	ctx := context.Background()
	a, err := s.CreateActivity(ctx, "学习")
	if err != nil {
		t.Fatalf("create activity: %v", err)
	}
	p, err := s.CreateProject(ctx, a.ID, "英语", 25)
	if err != nil {
		t.Fatalf("create project: %v", err)
	}
	return s, p
}

func TestSessionLifecycle(t *testing.T) {
	s, p := setupSessionFixture(t)
	ctx := context.Background()

	// No session running initially.
	if _, running, err := s.RunningSession(ctx); err != nil || running {
		t.Fatalf("expected no running session, got running=%v err=%v", running, err)
	}
	// Stop with nothing running.
	if _, err := s.StopSession(ctx, nil); !errors.Is(err, ErrNotRunning) {
		t.Fatalf("want ErrNotRunning, got %v", err)
	}

	sess, err := s.StartSession(ctx, p.ID, "背单词", 25)
	if err != nil {
		t.Fatalf("start: %v", err)
	}
	if sess.EndAt != nil {
		t.Fatalf("started session must have nil EndAt: %+v", sess)
	}
	if sess.ActivityName != "学习" || sess.ProjectName != "英语" {
		t.Fatalf("joined names wrong: %+v", sess)
	}

	// Running session is visible.
	run, running, err := s.RunningSession(ctx)
	if err != nil || !running || run.ID != sess.ID {
		t.Fatalf("running session wrong: running=%v sess=%+v err=%v", running, run, err)
	}

	// Second start is rejected.
	if _, err := s.StartSession(ctx, p.ID, "x", 0); !errors.Is(err, ErrAlreadyRunning) {
		t.Fatalf("want ErrAlreadyRunning, got %v", err)
	}

	// Stop with note override.
	note := "背完了 100 个"
	stopped, err := s.StopSession(ctx, &note)
	if err != nil {
		t.Fatalf("stop: %v", err)
	}
	if stopped.EndAt == nil || stopped.EndAt.Before(stopped.StartAt) {
		t.Fatalf("stopped session bad: %+v", stopped)
	}
	if stopped.Note != "背完了 100 个" {
		t.Fatalf("note override not applied: %q", stopped.Note)
	}

	// Nothing running after stop.
	if _, running, _ := s.RunningSession(ctx); running {
		t.Fatal("session should not be running after stop")
	}
	// Stop again.
	if _, err := s.StopSession(ctx, nil); !errors.Is(err, ErrNotRunning) {
		t.Fatalf("want ErrNotRunning, got %v", err)
	}
}

func TestStartSessionMissingProject(t *testing.T) {
	s, _ := setupSessionFixture(t)
	if _, err := s.StartSession(context.Background(), 999, "", 0); !errors.Is(err, ErrNotFound) {
		t.Fatalf("want ErrNotFound, got %v", err)
	}
}

// fixedZone builds a +08:00-like fixed zone for deterministic day tests.
var testZone = time.FixedZone("CST", 8*3600)

func TestListDaySessionsBoundaries(t *testing.T) {
	s, p := setupSessionFixture(t)
	ctx := context.Background()

	dayStart := time.Date(2026, 9, 16, 0, 0, 0, 0, testZone)
	dayEnd := time.Date(2026, 9, 17, 0, 0, 0, 0, testZone)

	insert := func(startUTC time.Time, endOffset time.Duration) {
		t.Helper()
		sess, err := s.StartSession(ctx, p.ID, "", 0)
		if err != nil {
			t.Fatalf("start: %v", err)
		}
		// Rewind the row's start_at directly to control the day bucket.
		_, err = s.db.Exec(`UPDATE sessions SET start_at = ?, end_at = ? WHERE id = ?`,
			fmtTime(startUTC), fmtTime(startUTC.Add(endOffset)), sess.ID)
		if err != nil {
			t.Fatalf("rewind: %v", err)
		}
	}

	// 23:50 local Sep 15 -> ends 00:10 Sep 16: belongs to Sep 15 (start day).
	insert(time.Date(2026, 9, 15, 23, 50, 0, 0, testZone), 20*time.Minute)
	// Exactly midnight Sep 16: belongs to Sep 16 (inclusive lower bound).
	insert(time.Date(2026, 9, 16, 0, 0, 0, 0, testZone), 30*time.Minute)
	// 14:00 Sep 16: belongs to Sep 16.
	insert(time.Date(2026, 9, 16, 14, 0, 0, 0, testZone), time.Hour)
	// Exactly midnight Sep 17: excluded from Sep 16 (exclusive upper bound).
	insert(time.Date(2026, 9, 17, 0, 0, 0, 0, testZone), time.Hour)

	got, err := s.ListDaySessions(ctx, dayStart.UTC(), dayEnd.UTC())
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("want 2 sessions on Sep 16, got %d: %+v", len(got), got)
	}
	wantStarts := []time.Time{
		time.Date(2026, 9, 16, 0, 0, 0, 0, testZone),
		time.Date(2026, 9, 16, 14, 0, 0, 0, testZone),
	}
	for i, sess := range got {
		if !sess.StartAt.Equal(wantStarts[i]) {
			t.Fatalf("session %d start %v, want %v", i, sess.StartAt, wantStarts[i])
		}
	}
	if !got[0].StartAt.Before(got[1].StartAt) {
		t.Fatal("sessions must be chronological")
	}

	// The cross-midnight session shows on Sep 15.
	sep15, err := s.ListDaySessions(ctx,
		time.Date(2026, 9, 15, 0, 0, 0, 0, testZone).UTC(),
		dayStart.UTC())
	if err != nil {
		t.Fatalf("list sep15: %v", err)
	}
	if len(sep15) != 1 {
		t.Fatalf("cross-midnight session must belong to Sep 15, got %d", len(sep15))
	}
}

func TestRunningSessionInDayList(t *testing.T) {
	s, p := setupSessionFixture(t)
	ctx := context.Background()

	if _, err := s.StartSession(ctx, p.ID, "进行中", 0); err != nil {
		t.Fatalf("start: %v", err)
	}
	now := time.Now()
	from, to := dayBoundsLocal(now)
	sessions, err := s.ListDaySessions(ctx, from, to)
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(sessions) != 1 || sessions[0].EndAt != nil {
		t.Fatalf("running session should appear in today's list with nil end: %+v", sessions)
	}
}

// dayBoundsLocal mirrors api.dayBounds for store-level tests.
func dayBoundsLocal(now time.Time) (time.Time, time.Time) {
	loc := now.Location()
	y, m, d := now.Date()
	return time.Date(y, m, d, 0, 0, 0, 0, loc).UTC(),
		time.Date(y, m, d+1, 0, 0, 0, 0, loc).UTC()
}
