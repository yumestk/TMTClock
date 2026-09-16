// Package store owns all SQLite persistence for ToMaToClock.
package store

import (
	"database/sql"
	_ "embed"
	"errors"
	"fmt"
	"strings"
	"time"

	_ "modernc.org/sqlite"
)

var (
	// ErrAlreadyRunning is returned when starting a session while one is running.
	ErrAlreadyRunning = errors.New("a session is already running")
	// ErrNotRunning is returned when stopping while no session is running.
	ErrNotRunning = errors.New("no session is running")
	// ErrHasSessions blocks deleting an activity or project that has recorded sessions.
	ErrHasSessions = errors.New("cannot delete: sessions exist")
	// ErrDuplicate is returned when a unique name constraint would be violated.
	ErrDuplicate = errors.New("duplicate name")
	// ErrNotFound is returned when updating or deleting a row that does not exist.
	ErrNotFound = errors.New("not found")
)

// Store wraps the SQLite database. All timestamps are stored as UTC,
// second-precision RFC3339 text so SQL string ranges stay chronological.
type Store struct {
	db *sql.DB
}

// Open opens (creating if needed) the database at path and applies the schema.
func Open(path string) (*Store, error) {
	dsn := fmt.Sprintf("file:%s?_pragma=busy_timeout(5000)&_pragma=journal_mode(WAL)&_pragma=foreign_keys(1)", path)
	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, err
	}
	s := &Store{db: db}
	if err := s.migrate(); err != nil {
		db.Close()
		return nil, err
	}
	return s, nil
}

func (s *Store) migrate() error {
	_, err := s.db.Exec(schemaSQL)
	return err
}

// Close closes the database.
func (s *Store) Close() error {
	return s.db.Close()
}

//go:embed schema.sql
var schemaSQL string

func fmtTime(t time.Time) string {
	return t.UTC().Truncate(time.Second).Format(time.RFC3339)
}

func parseTime(s string) (time.Time, error) {
	return time.Parse(time.RFC3339, s)
}

// isUnique reports whether err is a SQLite unique-constraint violation.
func isUnique(err error) bool {
	return err != nil && strings.Contains(err.Error(), "UNIQUE constraint failed")
}
