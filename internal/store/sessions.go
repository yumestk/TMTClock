package store

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"
)

// Session is a recorded time span. EndAt nil means the session is running —
// that IS the server-side timer state.
type Session struct {
	ID              int64      `json:"id"`
	ActivityName    string     `json:"activity"`
	ProjectName     string     `json:"project"`
	Note            string     `json:"note"`
	ExpectedMinutes int        `json:"expected_minutes"`
	StartAt         time.Time  `json:"start_at"`
	EndAt           *time.Time `json:"end_at"`
}

const sessionColumns = `
	SELECT s.id, a.name, p.name, s.note, s.expected_minutes, s.start_at, s.end_at
	FROM sessions s
	JOIN projects p ON p.id = s.project_id
	JOIN activities a ON a.id = p.activity_id`

func scanSession(row interface{ Scan(...any) error }) (Session, error) {
	var sess Session
	var start string
	var endNullable sql.NullString
	err := row.Scan(&sess.ID, &sess.ActivityName, &sess.ProjectName, &sess.Note,
		&sess.ExpectedMinutes, &start, &endNullable)
	if err != nil {
		return Session{}, err
	}
	sess.StartAt, err = parseTime(start)
	if err != nil {
		return Session{}, fmt.Errorf("parse start_at %q: %w", start, err)
	}
	if endNullable.Valid {
		t, err := parseTime(endNullable.String)
		if err != nil {
			return Session{}, fmt.Errorf("parse end_at %q: %w", endNullable.String, err)
		}
		sess.EndAt = &t
	}
	return sess, nil
}

// StartSession begins timing: inserts a session row with end_at NULL.
// Fails with ErrAlreadyRunning if another session is running, and
// ErrNotFound if the project does not exist.
func (s *Store) StartSession(ctx context.Context, projectID int64, note string, expectedMinutes int) (Session, error) {
	var sess Session
	err := s.withTx(ctx, func(tx *sql.Tx) error {
		var running int
		if err := tx.QueryRowContext(ctx, `SELECT COUNT(*) FROM sessions WHERE end_at IS NULL`).Scan(&running); err != nil {
			return err
		}
		if running > 0 {
			return ErrAlreadyRunning
		}
		var exists int
		if err := tx.QueryRowContext(ctx, `SELECT 1 FROM projects WHERE id = ?`, projectID).Scan(&exists); err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return fmt.Errorf("%w: project %d", ErrNotFound, projectID)
			}
			return err
		}
		start := fmtTime(time.Now())
		res, err := tx.ExecContext(ctx,
			`INSERT INTO sessions (project_id, note, expected_minutes, start_at, end_at) VALUES (?, ?, ?, ?, NULL)`,
			projectID, note, expectedMinutes, start)
		if err != nil {
			return err
		}
		id, err := res.LastInsertId()
		if err != nil {
			return err
		}
		sess, err = querySessionTx(ctx, tx, id)
		return err
	})
	if err != nil {
		return Session{}, err
	}
	return sess, nil
}

// StopSession ends the running session, optionally overriding its note.
func (s *Store) StopSession(ctx context.Context, note *string) (Session, error) {
	var sess Session
	err := s.withTx(ctx, func(tx *sql.Tx) error {
		var id int64
		err := tx.QueryRowContext(ctx, `SELECT id FROM sessions WHERE end_at IS NULL`).Scan(&id)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return ErrNotRunning
			}
			return err
		}
		end := fmtTime(time.Now())
		if _, err := tx.ExecContext(ctx,
			`UPDATE sessions SET end_at = ?, note = COALESCE(?, note) WHERE id = ?`,
			end, note, id); err != nil {
			return err
		}
		sess, err = querySessionTx(ctx, tx, id)
		return err
	})
	if err != nil {
		return Session{}, err
	}
	return sess, nil
}

// RunningSession returns the open session, if any.
func (s *Store) RunningSession(ctx context.Context) (Session, bool, error) {
	sess, err := scanFirst(s.db.QueryRowContext(ctx, sessionColumns+` WHERE s.end_at IS NULL`))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Session{}, false, nil
		}
		return Session{}, false, err
	}
	return sess, true, nil
}

// ListDaySessions returns sessions whose start_at falls in [from, to),
// in chronological order. The store is timezone-agnostic: callers pass
// explicit UTC bounds and own the day-boundary policy.
func (s *Store) ListDaySessions(ctx context.Context, from, to time.Time) ([]Session, error) {
	rows, err := s.db.QueryContext(ctx, sessionColumns+`
		WHERE s.start_at >= ? AND s.start_at < ?
		ORDER BY s.start_at`, fmtTime(from), fmtTime(to))
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var sessions []Session
	for rows.Next() {
		sess, err := scanSession(rows)
		if err != nil {
			return nil, err
		}
		sessions = append(sessions, sess)
	}
	return sessions, rows.Err()
}

func querySessionTx(ctx context.Context, tx *sql.Tx, id int64) (Session, error) {
	row := tx.QueryRowContext(ctx, sessionColumns+` WHERE s.id = ?`, id)
	return scanFirst(row)
}

func scanFirst(row *sql.Row) (Session, error) {
	return scanSession(row)
}

func (s *Store) withTx(ctx context.Context, fn func(tx *sql.Tx) error) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	if err := fn(tx); err != nil {
		tx.Rollback()
		return err
	}
	return tx.Commit()
}
