// Package main shows a client that walks every page of a paginated replify API,
// using UnwrapJSON and the pagination metadata to know when to stop.
//
// Run: go run ./examples/18_paginated_client
// Expected: three pages (items 1-3, 4-6, 7-8) and "total fetched: 8".
package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strconv"
	"time"

	"github.com/polarixa/replify"
	"github.com/polarixa/replify/pkg/slogger"
)

const (
	totalItems = 8
	perPage    = 3
)

func items(rw http.ResponseWriter, r *http.Request) {
	page, err := strconv.Atoi(r.URL.Query().Get("page"))
	if err != nil || page < 1 {
		page = 1
	}
	var data []int
	for i := (page-1)*perPage + 1; i <= min(page*perPage, totalItems); i++ {
		data = append(data, i)
	}
	replify.New().OK().
		WithBody(data).
		WithPagination(replify.FromPages(totalItems, perPage).WithPage(page)).
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
	srv := httptest.NewServer(http.HandlerFunc(items))
	defer srv.Close()
	client := &http.Client{Timeout: 5 * time.Second}

	fetched := 0
	for page := 1; ; page++ {
		resp, err := client.Get(fmt.Sprintf("%s/items?page=%d", srv.URL, page))
		if err != nil {
			return fmt.Errorf("page %d: %w", page, err)
		}
		raw, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			return fmt.Errorf("read page %d: %w", page, err)
		}
		w, err := replify.UnwrapJSON(string(raw))
		if err != nil {
			return fmt.Errorf("unwrap page %d: %w", page, err)
		}
		p := w.Pagination()
		slogger.Info("page fetched", slogger.Int("page", p.Page()), slogger.Int("total_pages", p.TotalPages()), slogger.String("items", w.BodyString()))
		var got []int
		if err := json.Unmarshal([]byte(w.BodyString()), &got); err != nil {
			return fmt.Errorf("decode page %d: %w", page, err)
		}
		fetched += len(got)
		if p.IsLast() {
			break
		}
	}
	slogger.Info("done", slogger.Int("total_fetched", fetched))
	return nil
}
