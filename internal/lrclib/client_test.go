package lrclib

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestTrack_HasSyncedLyrics(t *testing.T) {
	tests := []struct {
		name string
		track Track
		want bool
	}{
		{"has synced", Track{SyncedLyrics: "[00:00.00]Hello"}, true},
		{"empty synced", Track{SyncedLyrics: ""}, false},
		{"no brackets", Track{SyncedLyrics: "Hello"}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.track.HasSyncedLyrics(); got != tt.want {
				t.Errorf("HasSyncedLyrics() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTrack_HasPlainLyrics(t *testing.T) {
	tests := []struct {
		name string
		track Track
		want bool
	}{
		{"has plain", Track{PlainLyrics: "Hello\nWorld"}, true},
		{"empty plain", Track{PlainLyrics: ""}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.track.HasPlainLyrics(); got != tt.want {
				t.Errorf("HasPlainLyrics() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTrack_FormatDuration(t *testing.T) {
	tests := []struct {
		name string
		track Track
		want string
	}{
		{"normal", Track{Duration: 185.0}, "3:05"},
		{"zero", Track{Duration: 0}, "?:??"},
		{"one minute", Track{Duration: 60.0}, "1:00"},
		{"under minute", Track{Duration: 45.0}, "0:45"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.track.FormatDuration(); got != tt.want {
				t.Errorf("FormatDuration() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTrack_DisplayName(t *testing.T) {
	track := Track{ArtistName: "Artist", TrackName: "Song"}
	want := "Artist - Song"
	if got := track.DisplayName(); got != want {
		t.Errorf("DisplayName() = %v, want %v", got, want)
	}
}

func TestClient_Search(t *testing.T) {
	// Create test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/search" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		if r.URL.Query().Get("q") != "test" {
			t.Errorf("unexpected query: %s", r.URL.Query().Get("q"))
		}
		if r.Header.Get("User-Agent") == "" {
			t.Error("missing User-Agent header")
		}

		tracks := SearchResult{
			{ID: 1, TrackName: "Test Song", ArtistName: "Test Artist", Duration: 180},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(tracks)
	}))
	defer server.Close()

	client := &Client{
		httpClient: &http.Client{Timeout: 5 * time.Second},
		baseURL:    server.URL + "/api",
	}

	ctx := context.Background()
	tracks, err := client.Search(ctx, "test")
	if err != nil {
		t.Fatalf("Search() error: %v", err)
	}

	if len(tracks) != 1 {
		t.Errorf("expected 1 track, got %d", len(tracks))
	}
	if tracks[0].TrackName != "Test Song" {
		t.Errorf("unexpected track name: %s", tracks[0].TrackName)
	}
}

func TestClient_SearchByTrack(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/search" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		if r.URL.Query().Get("artist_name") != "Artist" {
			t.Errorf("unexpected artist: %s", r.URL.Query().Get("artist_name"))
		}
		if r.URL.Query().Get("track_name") != "Song" {
			t.Errorf("unexpected track: %s", r.URL.Query().Get("track_name"))
		}

		tracks := SearchResult{
			{ID: 1, TrackName: "Song", ArtistName: "Artist"},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(tracks)
	}))
	defer server.Close()

	client := &Client{
		httpClient: &http.Client{Timeout: 5 * time.Second},
		baseURL:    server.URL + "/api",
	}

	ctx := context.Background()
	tracks, err := client.SearchByTrack(ctx, "Artist", "Song")
	if err != nil {
		t.Fatalf("SearchByTrack() error: %v", err)
	}

	if len(tracks) != 1 {
		t.Errorf("expected 1 track, got %d", len(tracks))
	}
}

func TestClient_GetExact(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check path format: /get/Artist/Album/Track
		expectedPath := "/api/get/Artist/Album/Song"
		if r.URL.Path != expectedPath {
			t.Errorf("unexpected path: got %s, want %s", r.URL.Path, expectedPath)
		}

		track := Track{ID: 1, TrackName: "Song", ArtistName: "Artist", AlbumName: "Album"}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(track)
	}))
	defer server.Close()

	client := &Client{
		httpClient: &http.Client{Timeout: 5 * time.Second},
		baseURL:    server.URL + "/api",
	}

	ctx := context.Background()
	track, err := client.GetExact(ctx, "Artist", "Album", "Song")
	if err != nil {
		t.Fatalf("GetExact() error: %v", err)
	}

	if track == nil {
		t.Fatal("expected track, got nil")
	}
	if track.TrackName != "Song" {
		t.Errorf("unexpected track name: %s", track.TrackName)
	}
}

func TestClient_GetByID(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		expectedPath := "/api/get/123"
		if r.URL.Path != expectedPath {
			t.Errorf("unexpected path: got %s, want %s", r.URL.Path, expectedPath)
		}

		track := Track{ID: 123, TrackName: "Test", ArtistName: "Artist"}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(track)
	}))
	defer server.Close()

	client := &Client{
		httpClient: &http.Client{Timeout: 5 * time.Second},
		baseURL:    server.URL + "/api",
	}

	ctx := context.Background()
	track, err := client.GetByID(ctx, 123)
	if err != nil {
		t.Fatalf("GetByID() error: %v", err)
	}

	if track == nil {
		t.Fatal("expected track, got nil")
	}
	if track.ID != 123 {
		t.Errorf("unexpected ID: %d", track.ID)
	}
}

func TestClient_RateLimited(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusTooManyRequests)
		w.Write([]byte("rate limited"))
	}))
	defer server.Close()

	client := &Client{
		httpClient: &http.Client{Timeout: 5 * time.Second},
		baseURL:    server.URL + "/api",
	}

	ctx := context.Background()
	_, err := client.Search(ctx, "test")
	if err == nil {
		t.Fatal("expected error for rate limiting")
	}
}

func TestClient_Timeout(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(2 * time.Second)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := &Client{
		httpClient: &http.Client{Timeout: 100 * time.Millisecond},
		baseURL:    server.URL + "/api",
	}

	ctx := context.Background()
	_, err := client.Search(ctx, "test")
	if err == nil {
		t.Fatal("expected timeout error")
	}
}

func TestClient_NotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	client := &Client{
		httpClient: &http.Client{Timeout: 5 * time.Second},
		baseURL:    server.URL + "/api",
	}

	ctx := context.Background()
	tracks, err := client.Search(ctx, "nonexistent")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if tracks != nil {
		t.Errorf("expected nil tracks for 404, got %d", len(tracks))
	}
}

func TestNewClient(t *testing.T) {
	client := NewClient()
	if client == nil {
		t.Fatal("expected client, got nil")
	}
	if client.baseURL != baseURL {
		t.Errorf("baseURL = %s, want %s", client.baseURL, baseURL)
	}
}

func TestNewClientWithTimeout(t *testing.T) {
	timeout := 10 * time.Second
	client := NewClientWithTimeout(timeout)
	if client == nil {
		t.Fatal("expected client, got nil")
	}
	if client.httpClient.Timeout != timeout {
		t.Errorf("timeout = %v, want %v", client.httpClient.Timeout, timeout)
	}
}

func TestOptions_Defaults(t *testing.T) {
	opts := Options{}
	if opts.Query != "" {
		t.Error("default Query should be empty")
	}
	if opts.Artist != "" {
		t.Error("default Artist should be empty")
	}
	if opts.Track != "" {
		t.Error("default Track should be empty")
	}
	if opts.Output != "" {
		t.Error("default Output should be empty")
	}
	if opts.OffsetMs != 0 {
		t.Error("default OffsetMs should be 0")
	}
	if opts.Interactive != false {
		t.Error("default Interactive should be false")
	}
	if opts.PlainOnly != false {
		t.Error("default PlainOnly should be false")
	}
}

// --- Integration tests (skipped with -short) ---

func TestClient_Search_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	client := NewClient()
	ctx := context.Background()

	tracks, err := client.Search(ctx, "never gonna give you up")
	if err != nil {
		t.Fatalf("Search() error: %v", err)
	}

	if len(tracks) == 0 {
		t.Fatal("expected to find tracks")
	}

	// Check for Rick Astley's track
	found := false
	for _, track := range tracks {
		if strings.Contains(strings.ToLower(track.ArtistName), "rick astley") {
			found = true
			break
		}
	}
	if !found {
		t.Log("Rick Astley track not found (API results may vary)")
	}
}

func TestClient_GetExact_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	client := NewClient()
	ctx := context.Background()

	track, err := client.GetExact(ctx, "Rick Astley", "", "Never Gonna Give You Up")
	if err != nil {
		t.Fatalf("GetExact() error: %v", err)
	}

	if track == nil {
		t.Skip("track not found (API results may vary)")
	}

	if !track.HasSyncedLyrics() && !track.HasPlainLyrics() {
		t.Error("expected track to have lyrics")
	}
}
