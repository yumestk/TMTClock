# ToMaToClock

A local time-logging tool — a personal replacement for 番茄ToDo-style apps.
Start the timer, stop it, and the span is logged. No pomodoro cycles, no
accounts, no cloud: one binary, one SQLite file, your data.

- **Count-up timer** with optional expected duration — at the mark it chimes
  and posts a notification, but the timer keeps running until you stop it.
- **Server-held state**: a refresh or restart never loses the running
  session (it is a row in SQLite with `end_at IS NULL`).
- **Day view**: today's sessions as a timeline with proportional bars,
  expected-duration tick marks, and a running total.
- **Three-level model**: activity (分类) → project → session (with a note).

## Run the release build

Requires Go 1.22+ and Node 22 (see `.nvmrc`).

```sh
make build          # vite build → go:embed → ./tmtclock
./tmtclock          # opens the browser at http://127.0.0.1:8642
```

The database lives at `~/Library/Application Support/ToMaToClock/tmtclock.sqlite`
(macOS default; follows the OS user-config dir elsewhere). Override with:

```sh
./tmtclock -db /path/to/tmtclock.sqlite -addr 127.0.0.1:8642
```

Flags: `-db` (database file), `-addr` (listen address), `-no-browser`
(skip auto-open; used by `make dev-go`).

## Develop

```sh
make dev        # Go API on :8642 + Vite dev server on :5173 (proxies /api)
make test       # go test ./...
make build      # production build into ./tmtclock
make clean      # remove build output
```

Dev flow: edit `frontend/src/**` (HMR) or Go code (restart `dev-go`).
`make dev` uses `.nvmrc`'s Node via `check-node`; if your default node is
older, run `nvm use` first.

## Layout

```
cmd/tmtclock/    main: flags, wiring, embedded static files, open browser
internal/store/  all SQLite knowledge (schema, activities, projects, sessions)
internal/api/    thin HTTP layer: decode → store → encode (11 JSON routes)
frontend/        Vite + React + TS app; also a Go package embedding dist/
```

Time is stored as UTC, second-precision RFC3339 text; day boundaries are
computed in the server's local timezone. Sessions belong to the day they
started (cross-midnight sessions count to the start day). Activities and
projects with recorded sessions cannot be deleted — a time log must not
silently lose records.
