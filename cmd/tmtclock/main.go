// Command tmtclock runs the ToMaToClock local time-log server.
package main

import (
	"flag"
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/yumestk/TMTClock/frontend"
	"github.com/yumestk/TMTClock/internal/api"
	"github.com/yumestk/TMTClock/internal/store"
)

func main() {
	dbPath := flag.String("db", defaultDBPath(), "path to the SQLite database file")
	addr := flag.String("addr", "127.0.0.1:8642", "listen address")
	noBrowser := flag.Bool("no-browser", false, "do not open the browser on startup")
	flag.Parse()

	if err := os.MkdirAll(filepath.Dir(*dbPath), 0o755); err != nil {
		log.Fatalf("create db directory: %v", err)
	}
	s, err := store.Open(*dbPath)
	if err != nil {
		log.Fatalf("open store: %v", err)
	}
	defer s.Close()

	mux := api.New(s)
	dist, err := fs.Sub(frontend.Dist, "dist")
	if err != nil {
		log.Fatalf("embed dist: %v", err)
	}
	mux.Handle("GET /", http.FileServerFS(dist))

	url := "http://" + *addr
	srv := &http.Server{
		Addr:              *addr,
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
	}
	fmt.Printf("ToMaToClock listening on %s (db: %s)\n", url, *dbPath)
	if !*noBrowser {
		openBrowser(url)
	}
	if err := srv.ListenAndServe(); err != nil {
		log.Fatalf("serve: %v (is another tmtclock already running?)", err)
	}
}

func defaultDBPath() string {
	base, err := os.UserConfigDir()
	if err != nil {
		log.Fatalf("locate user config dir: %v", err)
	}
	return filepath.Join(base, "ToMaToClock", "tmtclock.sqlite")
}

func openBrowser(url string) {
	// Errors are non-fatal: the user can open the URL manually.
	_ = exec.Command("open", url).Start()
}
