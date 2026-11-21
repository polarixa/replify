// Package main demonstrates writing a replify response from an http.Handler:
// pagination, total, request metadata and hypermedia links.
//
// Run: go run ./examples/02_http_handler
// Expected: HTTP status line followed by the JSON envelope returned by /api/v2/users.
package main

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"time"

	"github.com/polarixa/replify"
	"github.com/polarixa/replify/pkg/slogger"
)

func listUsers(rw http.ResponseWriter, r *http.Request) {
	users := []map[string]string{
		{"id": "1", "name": "Alice"},
		{"id": "2", "name": "Bob"},
	}

	ls := replify.Links().
		WithSelf("/api/v2/users").
		WithNext("/api/v2/users?page=2").
		WithLink("create", "/api/v2/users", "POST").
		WithLinkObject("admin",
			replify.Link("/admin", "GET").
				WithTitle("Admin Panel").
				WithType("application/json"),
		)

	replify.New().
		OK().
		WithBody(users).
		WithPagination(replify.FromPages(100, 10)).
		WithLinks(ls).
		WithRequest(r).
		WithMessage("Users retrieved successfully").
		WriteJSON(rw)
}

func main() {
	// Development logger: text formatter on stdout. Use slogger.NewJSONFormatter() in production.
	slogger.SetGlobalLogger(slogger.New(
		slogger.WithOutput(os.Stdout),
		slogger.WithLevel(slogger.DebugLevel),
	))
	if err := run(); err != nil {
		slogger.Fatal("example failed", slogger.Err(err))
	}
}

func run() error {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/v2/users", listUsers)
	srv := httptest.NewServer(mux)
	defer srv.Close()

	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Get(srv.URL + "/api/v2/users")
	if err != nil {
		return fmt.Errorf("get users: %w", err)
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("read body: %w", err)
	}
	slogger.Info("response received", slogger.Int("status", resp.StatusCode), slogger.String("body", string(data)))
	return nil
}
