// Package main is a realistic mini "users" REST API built with replify:
// create (201), read (200/404), validation (422), conflict (409), list with
// pagination, and delete (204). It runs in-process against an in-memory store.
//
// Run: go run ./examples/15_rest_crud_api
// Expected: one "METHOD path -> status" line per call, with the key message or data.
package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/polarixa/replify"
	"github.com/polarixa/replify/pkg/slogger"
)

type user struct {
	ID    string `json:"id"`
	Email string `json:"email"`
}

type store struct {
	mu    sync.Mutex
	users map[string]user
}

func (s *store) handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("POST /users", s.create)
	mux.HandleFunc("GET /users", s.list)
	mux.HandleFunc("GET /users/{id}", s.get)
	mux.HandleFunc("DELETE /users/{id}", s.remove)
	return mux
}

func (s *store) create(rw http.ResponseWriter, r *http.Request) {
	var in user
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		replify.New().BadRequest().WithErrorf("invalid JSON: %v", err).WithRequest(r).WriteJSON(rw)
		return
	}
	if !strings.Contains(in.Email, "@") {
		replify.New().UnprocessableEntity().WithError("email must be a valid address").WithRequest(r).WriteJSON(rw)
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, u := range s.users {
		if u.Email == in.Email {
			replify.New().Conflict().WithError("email already registered").WithRequest(r).WriteJSON(rw)
			return
		}
	}
	in.ID = "u_" + strconv.Itoa(len(s.users)+1)
	s.users[in.ID] = in
	replify.New().Created().WithMessage("User created").WithBody(in).WithRequest(r).WriteJSON(rw)
}

func (s *store) get(rw http.ResponseWriter, r *http.Request) {
	s.mu.Lock()
	u, ok := s.users[r.PathValue("id")]
	s.mu.Unlock()
	if !ok {
		replify.New().NotFound().WithError("user not found").WithRequest(r).WriteJSON(rw)
		return
	}
	replify.New().OK().WithBody(u).WithRequest(r).WriteJSON(rw)
}

func (s *store) list(rw http.ResponseWriter, r *http.Request) {
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page < 1 {
		page = 1
	}
	const perPage = 2

	s.mu.Lock()
	all := make([]user, 0, len(s.users))
	for _, u := range s.users {
		all = append(all, u)
	}
	s.mu.Unlock()
	sort.Slice(all, func(i, j int) bool { return all[i].ID < all[j].ID })

	start := min((page-1)*perPage, len(all))
	end := min(start+perPage, len(all))
	replify.New().OK().
		WithBody(all[start:end]).
		WithPagination(replify.FromPages(len(all), perPage).WithPage(page)).
		WithRequest(r).
		WriteJSON(rw)
}

func (s *store) remove(rw http.ResponseWriter, r *http.Request) {
	s.mu.Lock()
	_, ok := s.users[r.PathValue("id")]
	delete(s.users, r.PathValue("id"))
	s.mu.Unlock()
	if !ok {
		replify.New().NotFound().WithError("user not found").WithRequest(r).WriteJSON(rw)
		return
	}
	replify.New().NoContent().WriteJSON(rw)
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

func call(c *http.Client, method, url, body string) error {
	req, err := http.NewRequest(method, url, bytes.NewBufferString(body))
	if err != nil {
		return err
	}
	resp, err := c.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	summary := ""
	if len(data) > 0 {
		w, err := replify.UnwrapJSON(string(data))
		if err != nil {
			return fmt.Errorf("unwrap: %w", err)
		}
		summary = w.BodyString()
		if w.IsClientError() {
			// The error text travels in the "issue" object of the envelope.
			var env struct {
				Issue struct {
					Message string `json:"message"`
				} `json:"issue"`
			}
			if err := json.Unmarshal(data, &env); err != nil {
				return fmt.Errorf("decode issue: %w", err)
			}
			summary = env.Issue.Message
		}
	}
	slogger.Info("call", slogger.String("method", method), slogger.String("uri", req.URL.RequestURI()), slogger.Int("status", resp.StatusCode), slogger.String("summary", summary))
	return nil
}

func run() error {
	srv := httptest.NewServer((&store{users: map[string]user{}}).handler())
	defer srv.Close()
	c := &http.Client{Timeout: 5 * time.Second}

	steps := []struct{ method, path, body string }{
		{"POST", "/users", `{"email":"a@example.com"}`},
		{"POST", "/users", `{"email":"b@example.com"}`},
		{"POST", "/users", `{"email":"c@example.com"}`},
		{"POST", "/users", `{"email":"a@example.com"}`},
		{"POST", "/users", `{"email":"nope"}`},
		{"POST", "/users", `{bad`},
		{"GET", "/users/u_1", ""},
		{"GET", "/users/u_99", ""},
		{"GET", "/users?page=2", ""},
		{"DELETE", "/users/u_1", ""},
		{"DELETE", "/users/u_1", ""},
	}
	for _, s := range steps {
		if err := call(c, s.method, srv.URL+s.path, s.body); err != nil {
			return fmt.Errorf("%s %s: %w", s.method, s.path, err)
		}
	}
	return nil
}
