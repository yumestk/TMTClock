package store

import (
	"context"
	"errors"
	"testing"
)

func TestProjectCRUD(t *testing.T) {
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
	if p.ID == 0 || p.ActivityID != a.ID || p.Name != "英语" || p.ExpectedMinutes != 25 {
		t.Fatalf("unexpected project: %+v", p)
	}

	acts, err := s.ListActivities(ctx)
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(acts) != 1 || len(acts[0].Projects) != 1 || acts[0].Projects[0].Name != "英语" {
		t.Fatalf("project not nested under activity: %+v", acts)
	}

	newName, newExpected := "日语", 45
	if err := s.UpdateProject(ctx, p.ID, &newName, &newExpected); err != nil {
		t.Fatalf("update: %v", err)
	}
	got, err := s.GetProject(ctx, p.ID)
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if got.Name != "日语" || got.ExpectedMinutes != 45 {
		t.Fatalf("update not applied: %+v", got)
	}

	if err := s.DeleteProject(ctx, p.ID); err != nil {
		t.Fatalf("delete: %v", err)
	}
	if _, err := s.GetProject(ctx, p.ID); !errors.Is(err, ErrNotFound) {
		t.Fatalf("want ErrNotFound after delete, got %v", err)
	}
}

func TestProjectPartialUpdate(t *testing.T) {
	s := newTestStore(t)
	ctx := context.Background()

	a, _ := s.CreateActivity(ctx, "学习")
	p, _ := s.CreateProject(ctx, a.ID, "英语", 25)

	// Omit expected_minutes: only name changes.
	name := "法语"
	if err := s.UpdateProject(ctx, p.ID, &name, nil); err != nil {
		t.Fatalf("update name only: %v", err)
	}
	got, _ := s.GetProject(ctx, p.ID)
	if got.Name != "法语" || got.ExpectedMinutes != 25 {
		t.Fatalf("partial update wrong: %+v", got)
	}

	// Omit name: only expected changes.
	expected := 60
	if err := s.UpdateProject(ctx, p.ID, nil, &expected); err != nil {
		t.Fatalf("update expected only: %v", err)
	}
	got, _ = s.GetProject(ctx, p.ID)
	if got.Name != "法语" || got.ExpectedMinutes != 60 {
		t.Fatalf("partial update wrong: %+v", got)
	}
}

func TestProjectDuplicateWithinActivity(t *testing.T) {
	s := newTestStore(t)
	ctx := context.Background()

	a, _ := s.CreateActivity(ctx, "学习")
	b, _ := s.CreateActivity(ctx, "工作")
	if _, err := s.CreateProject(ctx, a.ID, "英语", 0); err != nil {
		t.Fatalf("create: %v", err)
	}
	if _, err := s.CreateProject(ctx, a.ID, "英语", 0); !errors.Is(err, ErrDuplicate) {
		t.Fatalf("want ErrDuplicate within same activity, got %v", err)
	}
	// Same name under a different activity is allowed.
	if _, err := s.CreateProject(ctx, b.ID, "英语", 0); err != nil {
		t.Fatalf("same name in other activity should be allowed: %v", err)
	}
}

func TestProjectNotFoundAndInvalidActivity(t *testing.T) {
	s := newTestStore(t)
	ctx := context.Background()

	if _, err := s.CreateProject(ctx, 999, "英语", 0); !errors.Is(err, ErrNotFound) {
		t.Fatalf("want ErrNotFound for missing activity, got %v", err)
	}
	if err := s.UpdateProject(ctx, 999, nil, nil); !errors.Is(err, ErrNotFound) {
		t.Fatalf("want ErrNotFound, got %v", err)
	}
	if err := s.DeleteProject(ctx, 999); !errors.Is(err, ErrNotFound) {
		t.Fatalf("want ErrNotFound, got %v", err)
	}
}

func TestActivityDeleteCascadesEmptyProjects(t *testing.T) {
	s := newTestStore(t)
	ctx := context.Background()

	a, _ := s.CreateActivity(ctx, "学习")
	if _, err := s.CreateProject(ctx, a.ID, "英语", 0); err != nil {
		t.Fatalf("create project: %v", err)
	}
	// Projects have no sessions, so activity delete cascades them.
	if err := s.DeleteActivity(ctx, a.ID); err != nil {
		t.Fatalf("delete activity with empty projects: %v", err)
	}
	acts, _ := s.ListActivities(ctx)
	if len(acts) != 0 {
		t.Fatalf("expected empty list, got %+v", acts)
	}
}
