package kanye_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	kanye "github.com/tamnd/kanye-cli/kanye"
)

const fakeQuoteJSON = `{"quote":"Keep your nose out the business"}`

func newTestClient(ts *httptest.Server) *kanye.Client {
	cfg := kanye.DefaultConfig()
	cfg.BaseURL = ts.URL
	cfg.Rate = 0
	return kanye.NewClient(cfg)
}

func TestQuoteSendsUserAgent(t *testing.T) {
	var gotUA string
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotUA = r.Header.Get("User-Agent")
		_, _ = fmt.Fprint(w, fakeQuoteJSON)
	}))
	defer ts.Close()

	c := newTestClient(ts)
	_, err := c.Quote(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if gotUA == "" {
		t.Error("User-Agent not sent")
	}
}

func TestQuoteParses(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = fmt.Fprint(w, fakeQuoteJSON)
	}))
	defer ts.Close()

	c := newTestClient(ts)
	q, err := c.Quote(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if q.Quote != "Keep your nose out the business" {
		t.Errorf("Quote = %q", q.Quote)
	}
}

func TestQuotesCount(t *testing.T) {
	var calls int
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		json.NewEncoder(w).Encode(map[string]any{"quote": fmt.Sprintf("quote %d", calls)})
	}))
	defer ts.Close()

	c := newTestClient(ts)
	items, err := c.Quotes(context.Background(), 3)
	if err != nil {
		t.Fatal(err)
	}
	if len(items) != 3 {
		t.Errorf("len(items) = %d, want 3", len(items))
	}
	if calls != 3 {
		t.Errorf("server calls = %d, want 3", calls)
	}
}

func TestQuotesDefaultCount(t *testing.T) {
	var calls int
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		json.NewEncoder(w).Encode(map[string]any{"quote": "q"})
	}))
	defer ts.Close()

	c := newTestClient(ts)
	items, err := c.Quotes(context.Background(), 0)
	if err != nil {
		t.Fatal(err)
	}
	if len(items) != 5 {
		t.Errorf("len(items) = %d, want 5 (default)", len(items))
	}
	if calls != 5 {
		t.Errorf("server calls = %d, want 5", calls)
	}
}

func TestQuoteRetriesOn503(t *testing.T) {
	var hits int
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hits++
		if hits < 3 {
			w.WriteHeader(http.StatusServiceUnavailable)
			return
		}
		_, _ = fmt.Fprint(w, fakeQuoteJSON)
	}))
	defer ts.Close()

	cfg := kanye.DefaultConfig()
	cfg.BaseURL = ts.URL
	cfg.Rate = 0
	cfg.Retries = 3
	c := kanye.NewClient(cfg)

	_, err := c.Quote(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if hits != 3 {
		t.Errorf("server saw %d hits, want 3", hits)
	}
}
