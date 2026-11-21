// Package main shows custom authentication/authorization middleware that returns
// replify 401/403 envelopes, composed with the built-in Logger and Recovery.
//
// Run: go run ./examples/17_auth_middleware
// Expected: no token -> 401, user token on admin route -> 403, admin token -> 200.
package main

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"time"

	"github.com/polarixa/replify"
	"github.com/polarixa/replify/pkg/slogger"
)

var tokens = map[string]string{"user-token": "user", "admin-token": "admin"}

// authenticate rejects requests without a known bearer token.
func authenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		role, ok := tokens[r.Header.Get("Authorization")]
		if !ok {
			replify.New().Unauthorized().WithError("missing or invalid token").WithRequest(r).WriteJSON(rw)
			return
		}
		r.Header.Set("X-Role", role)
		next.ServeHTTP(rw, r)
	})
}

// requireRole rejects authenticated requests lacking the needed role.
func requireRole(role string, next http.Handler) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		if r.Header.Get("X-Role") != role {
			replify.New().Forbidden().WithError("insufficient role").WithRequest(r).WriteJSON(rw)
			return
		}
		next.ServeHTTP(rw, r)
	})
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
	mux.Handle("/admin", requireRole("admin", http.HandlerFunc(func(rw http.ResponseWriter, _ *http.Request) {
		replify.New().OK().WithMessage("welcome, admin").WriteJSON(rw)
	})))

	srv := httptest.NewServer(replify.Logger()(replify.Recovery()(authenticate(mux))))
	defer srv.Close()

	client := &http.Client{Timeout: 5 * time.Second}
	for _, token := range []string{"", "user-token", "admin-token"} {
		req, err := http.NewRequest(http.MethodGet, srv.URL+"/admin", nil)
		if err != nil {
			return err
		}
		req.Header.Set("Authorization", token)
		resp, err := client.Do(req)
		if err != nil {
			return fmt.Errorf("request: %w", err)
		}
		resp.Body.Close()
		slogger.Info("client received", slogger.String("token", token), slogger.Int("status", resp.StatusCode))
	}
	return nil
}
