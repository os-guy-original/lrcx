// Package lrclib provides a client for the LRCLib API.
package lrclib

// Track represents a track from LRCLib API.
type Track struct {
	ID            int     `json:"id"`
	TrackName     string  `json:"trackName"`
	ArtistName    string  `json:"artistName"`
	AlbumName     string  `json:"albumName"`
	Duration      float64 `json:"duration"`
	Instrumental  bool    `json:"instrumental"`
	PlainLyrics   string  `json:"plainLyrics"`
	SyncedLyrics  string  `json:"syncedLyrics"`
}

// SearchResult represents a search response from LRCLib.
type SearchResult []Track

// Options configures the LRCLib client.
type Options struct {
	Query       string
	Artist      string
	Track       string
	Album       string
	Output      string
	OffsetMs    int
	Interactive bool
	PlainOnly   bool // Get plain lyrics only (no timestamps)
	Verbose     bool // Show debug output
}
