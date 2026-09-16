package store

import (
	"context"
	"fmt"
	"time"
)

// Activity is a top-level category (e.g. 学习/工作/运动) with nested projects.
type Activity struct {
	ID       int64     `json:"id"`
	Name     string    `json:"name"`
	Projects []Project `json:"projects"`
}

// Project belongs to an activity; sessions always reference a project.
type Project struct {
	ID              int64  `json:"id"`
	ActivityID      int64  `json:"activity_id"`
	Name            string `json:"name"`
	ExpectedMinutes int    `json:"expected_minutes"`
}

// ListActivities returns all activities with their projects, ordered by name.
func (s *Store) ListActivities(ctx context.Context) ([]Activity, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT id, name FROM activities ORDER BY name`)
	if err != nil {
		return nil, err
	}
	var acts []Activity
	for rows.Next() {
		var a Activity
		if err := rows.Scan(&a.ID, &a.Name); err != nil {
			rows.Close()
			return nil, err
		}
		acts = append(acts, a)
	}
	rows.Close()
	if err := rows.Err(); err != nil {
		return nil, err
	}

	projRows, err := s.db.QueryContext(ctx, `SELECT id, activity_id, name, expected_minutes FROM projects ORDER BY name`)
	if err != nil {
		return nil, err
	}
	defer projRows.Close()
	byAct := map[int64][]Project{}
	for projRows.Next() {
		var p Project
		if err := projRows.Scan(&p.ID, &p.ActivityID, &p.Name, &p.ExpectedMinutes); err != nil {
			return nil, err
		}
		byAct[p.ActivityID] = append(byAct[p.ActivityID], p)
	}
	if err := projRows.Err(); err != nil {
		return nil, err
	}

	for i := range acts {
		acts[i].Projects = byAct[acts[i].ID]
	}
	return acts, nil
}

// CreateActivity inserts a new activity, returning it with its id.
func (s *Store) CreateActivity(ctx context.Context, name string) (Activity, error) {
	now := fmtTime(time.Now())
	res, err := s.db.ExecContext(ctx, `INSERT INTO activities (name, created_at) VALUES (?, ?)`, name, now)
	if err != nil {
		if isUnique(err) {
			return Activity{}, fmt.Errorf("%w: activity %q", ErrDuplicate, name)
		}
		return Activity{}, err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return Activity{}, err
	}
	return Activity{ID: id, Name: name}, nil
}

// RenameActivity updates an activity's name.
func (s *Store) RenameActivity(ctx context.Context, id int64, name string) error {
	res, err := s.db.ExecContext(ctx, `UPDATE activities SET name = ? WHERE id = ?`, name, id)
	if err != nil {
		if isUnique(err) {
			return fmt.Errorf("%w: activity %q", ErrDuplicate, name)
		}
		return err
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return ErrNotFound
	}
	return nil
}

// DeleteActivity removes an activity (cascading its projects) unless any of
// its projects has recorded sessions.
func (s *Store) DeleteActivity(ctx context.Context, id int64) error {
	var count int
	err := s.db.QueryRowContext(ctx, `
		SELECT COUNT(*) FROM sessions s
		JOIN projects p ON p.id = s.project_id
		WHERE p.activity_id = ?`, id).Scan(&count)
	if err != nil {
		return err
	}
	if count > 0 {
		return fmt.Errorf("%w: activity has %d session(s)", ErrHasSessions, count)
	}
	res, err := s.db.ExecContext(ctx, `DELETE FROM activities WHERE id = ?`, id)
	if err != nil {
		return err
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return ErrNotFound
	}
	return nil
}
