package store

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"
)

// CreateProject inserts a project under the given activity.
func (s *Store) CreateProject(ctx context.Context, activityID int64, name string, expectedMinutes int) (Project, error) {
	var exists int
	err := s.db.QueryRowContext(ctx, `SELECT 1 FROM activities WHERE id = ?`, activityID).Scan(&exists)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Project{}, fmt.Errorf("%w: activity %d", ErrNotFound, activityID)
		}
		return Project{}, err
	}
	if expectedMinutes < 0 {
		return Project{}, errors.New("expected_minutes must be >= 0")
	}
	now := fmtTime(time.Now())
	res, err := s.db.ExecContext(ctx,
		`INSERT INTO projects (activity_id, name, expected_minutes, created_at) VALUES (?, ?, ?, ?)`,
		activityID, name, expectedMinutes, now)
	if err != nil {
		if isUnique(err) {
			return Project{}, fmt.Errorf("%w: project %q under activity %d", ErrDuplicate, name, activityID)
		}
		return Project{}, err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return Project{}, err
	}
	return Project{ID: id, ActivityID: activityID, Name: name, ExpectedMinutes: expectedMinutes}, nil
}

// UpdateProject partially updates a project: nil pointers mean "leave unchanged".
func (s *Store) UpdateProject(ctx context.Context, id int64, name *string, expectedMinutes *int) error {
	var current Project
	err := s.db.QueryRowContext(ctx,
		`SELECT id, activity_id, name, expected_minutes FROM projects WHERE id = ?`, id).
		Scan(&current.ID, &current.ActivityID, &current.Name, &current.ExpectedMinutes)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return fmt.Errorf("%w: project %d", ErrNotFound, id)
		}
		return err
	}
	nextName := current.Name
	if name != nil {
		nextName = *name
	}
	nextExpected := current.ExpectedMinutes
	if expectedMinutes != nil {
		if *expectedMinutes < 0 {
			return errors.New("expected_minutes must be >= 0")
		}
		nextExpected = *expectedMinutes
	}
	if nextName == current.Name && nextExpected == current.ExpectedMinutes {
		return nil
	}
	_, err = s.db.ExecContext(ctx,
		`UPDATE projects SET name = ?, expected_minutes = ? WHERE id = ?`,
		nextName, nextExpected, id)
	if err != nil {
		if isUnique(err) {
			return fmt.Errorf("%w: project %q under activity %d", ErrDuplicate, nextName, current.ActivityID)
		}
		return err
	}
	return nil
}

// DeleteProject removes a project unless it has recorded sessions.
func (s *Store) DeleteProject(ctx context.Context, id int64) error {
	var count int
	err := s.db.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM sessions WHERE project_id = ?`, id).Scan(&count)
	if err != nil {
		return err
	}
	if count > 0 {
		return fmt.Errorf("%w: project has %d session(s)", ErrHasSessions, count)
	}
	res, err := s.db.ExecContext(ctx, `DELETE FROM projects WHERE id = ?`, id)
	if err != nil {
		return err
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return fmt.Errorf("%w: project %d", ErrNotFound, id)
	}
	return nil
}

// GetProject fetches a single project by id.
func (s *Store) GetProject(ctx context.Context, id int64) (Project, error) {
	var p Project
	err := s.db.QueryRowContext(ctx,
		`SELECT id, activity_id, name, expected_minutes FROM projects WHERE id = ?`, id).
		Scan(&p.ID, &p.ActivityID, &p.Name, &p.ExpectedMinutes)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Project{}, fmt.Errorf("%w: project %d", ErrNotFound, id)
		}
		return Project{}, err
	}
	return p, nil
}
