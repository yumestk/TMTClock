package store

import (
	"context"
	"errors"
	"path/filepath"
	"testing"
)

func newTestStore(t *testing.T) *Store {
	t.Helper()
	s, err := Open(filepath.Join(t.TempDir(), "test.sqlite"))
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	t.Cleanup(func() { s.Close() })
	return s
}

func TestActivityCRUD(t *testing.T) {
	s := newTestStore(t)
	ctx := context.Background()

	a, err := s.CreateActivity(ctx, "学习")
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	if a.ID == 0 || a.Name != "学习" {
		t.Fatalf("unexpected activity: %+v", a)
	}

	acts, err := s.ListActivities(ctx)
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(acts) != 1 || acts[0].Name != "学习" {
		t.Fatalf("unexpected list: %+v", acts)
	}

	if err := s.RenameActivity(ctx, a.ID, "工作"); err != nil {
		t.Fatalf("rename: %v", err)
	}
	acts, _ = s.ListActivities(ctx)
	if acts[0].Name != "工作" {
		t.Fatalf("rename not applied: %+v", acts)
	}

	if err := s.DeleteActivity(ctx, a.ID); err != nil {
		t.Fatalf("delete: %v", err)
	}
	acts, _ = s.ListActivities(ctx)
	if len(acts) != 0 {
		t.Fatalf("expected empty list, got %+v", acts)
	}
}

func TestActivityDuplicate(t *testing.T) {
	s := newTestStore(t)
	ctx := context.Background()

	if _, err := s.CreateActivity(ctx, "学习"); err != nil {
		t.Fatalf("create: %v", err)
	}
	_, err := s.CreateActivity(ctx, "学习")
	if !errors.Is(err, ErrDuplicate) {
		t.Fatalf("want ErrDuplicate, got %v", err)
	}
}

func TestActivityNotFound(t *testing.T) {
	s := newTestStore(t)
	ctx := context.Background()

	if err := s.RenameActivity(ctx, 999, "x"); !errors.Is(err, ErrNotFound) {
		t.Fatalf("want ErrNotFound, got %v", err)
	}
	if err := s.DeleteActivity(ctx, 999); !errors.Is(err, ErrNotFound) {
		t.Fatalf("want ErrNotFound, got %v", err)
	}
}
