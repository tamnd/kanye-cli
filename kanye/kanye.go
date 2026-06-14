// Package kanye is the library behind the kanye command line:
// the HTTP client, request shaping, and the typed data models for
// the Kanye West quotes API (api.kanye.rest).
package kanye

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"
)

// Host is the site this client talks to.
const Host = "api.kanye.rest"

// Config holds all tunable parameters for the Client.
type Config struct {
	BaseURL   string
	UserAgent string
	Rate      time.Duration
	Timeout   time.Duration
	Retries   int
}

// DefaultConfig returns a Config with sensible defaults.
func DefaultConfig() Config {
	return Config{
		BaseURL:   "https://api.kanye.rest",
		UserAgent: "kanye-cli/0.1.0 (github.com/tamnd/kanye-cli)",
		Rate:      200 * time.Millisecond,
		Timeout:   30 * time.Second,
		Retries:   3,
	}
}

// Client talks to api.kanye.rest over HTTP.
type Client struct {
	cfg  Config
	http *http.Client
	mu   sync.Mutex
	last time.Time
}

// NewClient returns a Client configured with cfg.
func NewClient(cfg Config) *Client {
	return &Client{
		cfg:  cfg,
		http: &http.Client{Timeout: cfg.Timeout},
	}
}

// Quote fetches one random Kanye West quote.
func (c *Client) Quote(ctx context.Context) (Quote, error) {
	u := c.cfg.BaseURL + "/"
	body, err := c.get(ctx, u)
	if err != nil {
		return Quote{}, err
	}
	var q quoteResponse
	if err := json.Unmarshal(body, &q); err != nil {
		return Quote{}, fmt.Errorf("decode quote: %w", err)
	}
	return Quote{Quote: q.Quote}, nil
}

// Quotes fetches count random Kanye West quotes by calling the API count times.
func (c *Client) Quotes(ctx context.Context, count int) ([]Quote, error) {
	if count <= 0 {
		count = 5
	}
	items := make([]Quote, 0, count)
	for i := 0; i < count; i++ {
		q, err := c.Quote(ctx)
		if err != nil {
			return nil, err
		}
		q.Rank = i + 1
		items = append(items, q)
	}
	return items, nil
}

func (c *Client) get(ctx context.Context, url string) ([]byte, error) {
	var lastErr error
	for attempt := 0; attempt <= c.cfg.Retries; attempt++ {
		if attempt > 0 {
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(backoff(attempt)):
			}
		}
		body, retry, err := c.do(ctx, url)
		if err == nil {
			return body, nil
		}
		lastErr = err
		if !retry {
			return nil, err
		}
	}
	return nil, fmt.Errorf("get %s: %w", url, lastErr)
}

func (c *Client) do(ctx context.Context, rawURL string) ([]byte, bool, error) {
	c.pace()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, rawURL, nil)
	if err != nil {
		return nil, false, err
	}
	req.Header.Set("User-Agent", c.cfg.UserAgent)
	resp, err := c.http.Do(req)
	if err != nil {
		return nil, true, err
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode == http.StatusTooManyRequests || resp.StatusCode >= 500 {
		return nil, true, fmt.Errorf("http %d", resp.StatusCode)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, false, fmt.Errorf("http %d", resp.StatusCode)
	}
	b, err := io.ReadAll(resp.Body)
	return b, err != nil, err
}

func (c *Client) pace() {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.cfg.Rate <= 0 {
		return
	}
	if wait := c.cfg.Rate - time.Since(c.last); wait > 0 {
		time.Sleep(wait)
	}
	c.last = time.Now()
}

func backoff(attempt int) time.Duration {
	return min(time.Duration(attempt)*500*time.Millisecond, 5*time.Second)
}

// Quote is a single Kanye West quote.
type Quote struct {
	Rank  int    `json:"rank,omitempty"`
	Quote string `json:"quote"`
}

// quoteResponse is the raw JSON from api.kanye.rest.
type quoteResponse struct {
	Quote string `json:"quote"`
}
