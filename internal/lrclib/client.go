package lrclib

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

const (
	baseURL    = "https://lrclib.net/api"
	userAgent  = "lrcx/0.2.0 (https://github.com/os-guy-original/lrcx)"
	defaultTimeout = 30 * time.Second
)

// Client provides methods to interact with the LRCLib API.
type Client struct {
	httpClient *http.Client
	baseURL    string
}

// NewClient creates a new LRCLib client.
func NewClient() *Client {
	return &Client{
		httpClient: &http.Client{
			Timeout: defaultTimeout,
		},
		baseURL: baseURL,
	}
}

// NewClientWithTimeout creates a new LRCLib client with custom timeout.
func NewClientWithTimeout(timeout time.Duration) *Client {
	return &Client{
		httpClient: &http.Client{
			Timeout: timeout,
		},
		baseURL: baseURL,
	}
}

// Search searches for tracks by query string.
func (c *Client) Search(ctx context.Context, query string) (SearchResult, error) {
	endpoint := fmt.Sprintf("%s/search?q=%s", c.baseURL, url.QueryEscape(query))
	return c.doRequest(ctx, endpoint)
}

// SearchByTrack searches for tracks by artist and track name.
func (c *Client) SearchByTrack(ctx context.Context, artist, track string) (SearchResult, error) {
	endpoint := fmt.Sprintf("%s/search?track_name=%s&artist_name=%s",
		c.baseURL, url.QueryEscape(track), url.QueryEscape(artist))
	return c.doRequest(ctx, endpoint)
}

// GetExact gets a track by exact artist, album, and track name.
// Album can be empty string if not needed.
func (c *Client) GetExact(ctx context.Context, artist, album, track string) (*Track, error) {
	endpoint := fmt.Sprintf("%s/get/%s/%s/%s",
		c.baseURL,
		url.PathEscape(artist),
		url.PathEscape(album),
		url.PathEscape(track))
	tracks, err := c.doRequest(ctx, endpoint)
	if err != nil {
		return nil, err
	}
	if len(tracks) == 0 {
		return nil, nil
	}
	return &tracks[0], nil
}

// GetByID gets a track by its ID.
func (c *Client) GetByID(ctx context.Context, id int) (*Track, error) {
	endpoint := fmt.Sprintf("%s/get/%d", c.baseURL, id)
	tracks, err := c.doRequest(ctx, endpoint)
	if err != nil {
		return nil, err
	}
	if len(tracks) == 0 {
		return nil, nil
	}
	return &tracks[0], nil
}

func (c *Client) doRequest(ctx context.Context, endpoint string) (SearchResult, error) {
	// Debug: print the endpoint
	if os.Getenv("LRCX_DEBUG") != "" {
		fmt.Fprintf(os.Stderr, "DEBUG: Requesting: %s\n", endpoint)
	}

	req, err := http.NewRequestWithContext(ctx, "GET", endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("User-Agent", userAgent)
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			return nil, fmt.Errorf("request timed out")
		}
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusTooManyRequests {
		return nil, fmt.Errorf("rate limited, please wait before making more requests")
	}

	if resp.StatusCode == http.StatusNotFound {
		return nil, nil
	}

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(body))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	// Try to parse as a single track first (for /get/ endpoint)
	var singleTrack Track
	if err := json.Unmarshal(body, &singleTrack); err == nil && singleTrack.ID != 0 {
		return SearchResult{singleTrack}, nil
	}

	// Try to parse as array of tracks
	var tracks SearchResult
	if err := json.Unmarshal(body, &tracks); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return tracks, nil
}

// HasSyncedLyrics returns true if the track has synchronized (LRC) lyrics.
func (t *Track) HasSyncedLyrics() bool {
	return t.SyncedLyrics != "" && strings.Contains(t.SyncedLyrics, "[")
}

// HasPlainLyrics returns true if the track has plain text lyrics.
func (t *Track) HasPlainLyrics() bool {
	return t.PlainLyrics != ""
}

// FormatDuration returns the duration in mm:ss format.
func (t *Track) FormatDuration() string {
	if t.Duration <= 0 {
		return "?:??"
	}
	duration := int(t.Duration)
	minutes := duration / 60
	seconds := duration % 60
	return fmt.Sprintf("%d:%02d", minutes, seconds)
}

// DisplayName returns a formatted string for display.
func (t *Track) DisplayName() string {
	return fmt.Sprintf("%s - %s", t.ArtistName, t.TrackName)
}
