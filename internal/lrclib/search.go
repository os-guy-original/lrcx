package lrclib

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/os-guy-original/lrcx/internal/ui"
)

// RunWithOpts searches and downloads lyrics with full options.
func RunWithOpts(opts Options) error {
	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()

	client := NewClient()

	var tracks SearchResult
	var err error

	// Debug: show search parameters
	if opts.Verbose {
		fmt.Fprintf(os.Stderr, "\nSearch params: artist=%q, track=%q, query=%q\n",
			opts.Artist, opts.Track, opts.Query)
	}

	stop := ui.Spin("Searching LRCLib", opts.Verbose)

	if opts.Artist != "" && opts.Track != "" {
		tracks, err = client.SearchByTrack(ctx, opts.Artist, opts.Track)
	} else if opts.Query != "" {
		tracks, err = client.Search(ctx, opts.Query)
	} else {
		stop(nil)
		return fmt.Errorf("please provide a search query or use --artist and --track flags")
	}

	stop(err)
	if err != nil {
		return fmt.Errorf("search failed: %w", err)
	}

	if len(tracks) == 0 {
		return fmt.Errorf("no lyrics found")
	}

	// Debug: show what we got
	if opts.Verbose {
		fmt.Fprintf(os.Stderr, "\nFound %d results:\n", len(tracks))
		for i, t := range tracks {
			fmt.Fprintf(os.Stderr, "  %d: %s - %s (instrumental: %v, has_synced: %v)\n",
				i+1, t.ArtistName, t.TrackName, t.Instrumental, t.HasSyncedLyrics())
		}
	}

	// Filter tracks by artist if artist was specified
	stop = ui.Spin("Filtering results", opts.Verbose)
	var filteredTracks SearchResult
	if opts.Artist != "" {
		for _, t := range tracks {
			if artistMatches(opts.Artist, t.ArtistName) {
				filteredTracks = append(filteredTracks, t)
			}
		}
	} else {
		filteredTracks = tracks
	}
	stop(nil)

	// Check if we found any matching tracks
	if len(filteredTracks) == 0 {
		// No exact match - warn about mismatched artists
		if opts.Artist != "" && len(tracks) > 0 {
			fmt.Fprintf(os.Stderr, "Warning: No exact match found for artist %q\n", opts.Artist)
			fmt.Fprintf(os.Stderr, "Found results for: %s\n", tracks[0].ArtistName)

			if !opts.Interactive {
				fmt.Fprintf(os.Stderr, "Use --interactive to select from results, or try a different search.\n")
				return fmt.Errorf("no lyrics found for artist %q", opts.Artist)
			}

			// In interactive mode, offer to show all results
			fmt.Fprintf(os.Stderr, "Would you like to see all results? [y/N]: ")
			reader := bufio.NewReader(os.Stdin)
			input, _ := reader.ReadString('\n')
			input = strings.TrimSpace(strings.ToLower(input))
			if input != "y" && input != "yes" {
				return fmt.Errorf("no lyrics found for artist %q", opts.Artist)
			}
			filteredTracks = tracks
		} else {
			return fmt.Errorf("no lyrics found")
		}
	}

	var selected *Track
	if opts.Interactive && len(filteredTracks) > 1 {
		selected = promptSelectTrack(filteredTracks)
	} else if opts.Interactive && len(filteredTracks) == 1 {
		// Show the single result and confirm
		t := filteredTracks[0]
		fmt.Fprintf(os.Stderr, "Found: %s - %s [%s]\n", t.ArtistName, t.TrackName, t.FormatDuration())
		fmt.Fprintf(os.Stderr, "Use this? [Y/n]: ")
		reader := bufio.NewReader(os.Stdin)
		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(strings.ToLower(input))
		if input == "n" || input == "no" {
			return fmt.Errorf("cancelled by user")
		}
		selected = &t
	} else {
		// Auto-select best match
		// Prefer tracks with synced lyrics that match artist
		for i := range filteredTracks {
			if filteredTracks[i].HasSyncedLyrics() {
				selected = &filteredTracks[i]
				break
			}
		}
		// Fall back to first track if none have synced lyrics
		if selected == nil {
			for i := range filteredTracks {
				if filteredTracks[i].HasPlainLyrics() {
					selected = &filteredTracks[i]
					break
				}
			}
		}
		if selected == nil {
			selected = &filteredTracks[0]
		}
	}

	// Warn if selected artist doesn't match (for non-interactive mode)
	if opts.Artist != "" && !artistMatches(opts.Artist, selected.ArtistName) && !opts.Interactive {
		fmt.Fprintf(os.Stderr, "Warning: Artist mismatch!\n")
		fmt.Fprintf(os.Stderr, "  Searched for: %s\n", opts.Artist)
		fmt.Fprintf(os.Stderr, "  Found: %s\n", selected.ArtistName)
		fmt.Fprintf(os.Stderr, "Use --interactive to confirm or select a different result.\n")
	}

	// Show what we found
	stop = ui.Spin("Fetching lyrics", opts.Verbose)

	// Get the lyrics
	lyrics := selected.SyncedLyrics
	if opts.PlainOnly || lyrics == "" {
		lyrics = selected.PlainLyrics
	}

	if lyrics == "" {
		stop(nil)
		if selected.Instrumental {
			return fmt.Errorf("this track is instrumental (no lyrics)")
		}
		return fmt.Errorf("no lyrics available for this track")
	}

	// Apply offset if needed
	if opts.OffsetMs != 0 {
		lyrics = applyOffset(lyrics, opts.OffsetMs)
	}

	stop(nil)

	// Show success message
	lyricsType := "synced"
	if opts.PlainOnly || selected.SyncedLyrics == "" {
		lyricsType = "plain"
	}
	fmt.Fprintf(os.Stderr, "✓ Found %s lyrics: %s - %s\n", lyricsType, selected.ArtistName, selected.TrackName)

	// Output
	var w *os.File
	if opts.Output == "" {
		w = os.Stdout
	} else {
		w, err = os.Create(opts.Output)
		if err != nil {
			return fmt.Errorf("cannot create output file: %w", err)
		}
		defer w.Close()
	}

	fmt.Fprintln(w, lyrics)
	return nil
}

// artistMatches checks if the search artist matches the result artist (case-insensitive, partial match)
func artistMatches(search, result string) bool {
	search = strings.ToLower(strings.TrimSpace(search))
	result = strings.ToLower(strings.TrimSpace(result))

	// Exact match
	if search == result {
		return true
	}

	// Partial match (search is contained in result or vice versa)
	if strings.Contains(result, search) || strings.Contains(search, result) {
		return true
	}

	// Check if main words match (ignoring "feat", "ft", etc.)
	searchWords := extractMainWords(search)
	resultWords := extractMainWords(result)

	for _, sw := range searchWords {
		found := false
		for _, rw := range resultWords {
			if sw == rw {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}

	return true
}

// extractMainWords extracts the main words from an artist name (removes "feat", "ft", etc.)
func extractMainWords(s string) []string {
	// Remove common separators and featuring markers
	s = strings.ToLower(s)
	for _, sep := range []string{" feat", " ft", " feat.", " ft.", " featuring", " & ", " x ", " vs "} {
		if idx := strings.Index(s, sep); idx > 0 {
			s = s[:idx]
		}
	}

	words := strings.Fields(s)
	var mainWords []string
	skipWords := map[string]bool{"the": true, "a": true, "an": true, "and": true}

	for _, w := range words {
		if !skipWords[w] && len(w) > 1 {
			mainWords = append(mainWords, w)
		}
	}

	return mainWords
}

func promptSelectTrack(tracks SearchResult) *Track {
	// Limit to max 15 results to avoid overwhelming the user
	maxResults := 15
	if len(tracks) > maxResults {
		tracks = tracks[:maxResults]
	}

	fmt.Fprintln(os.Stderr, "\nFound multiple matches:")
	for i, t := range tracks {
		syncMarker := ""
		if t.HasSyncedLyrics() {
			syncMarker = " ✓"
		} else if t.HasPlainLyrics() {
			syncMarker = " (plain)"
		} else if t.Instrumental {
			syncMarker = " (instrumental)"
		} else {
			syncMarker = " (no lyrics)"
		}
		fmt.Fprintf(os.Stderr, "  %2d) %s - %s [%s]%s\n", i+1, t.ArtistName, t.TrackName, t.FormatDuration(), syncMarker)
	}

	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Fprint(os.Stderr, "\nSelect track [1]: ")
		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(input)
		if input == "" {
			return &tracks[0]
		}
		var idx int
		if _, err := fmt.Sscanf(input, "%d", &idx); err == nil && idx >= 1 && idx <= len(tracks) {
			return &tracks[idx-1]
		}
		fmt.Fprintln(os.Stderr, "Invalid selection")
	}
}

// applyOffset adjusts all timestamps in LRC format by the given milliseconds.
func applyOffset(lrc string, offsetMs int) string {
	lines := strings.Split(lrc, "\n")
	var result []string

	for _, line := range lines {
		// Skip empty lines and metadata lines
		if line == "" || strings.HasPrefix(line, "[") && !isTimestampLine(line) {
			result = append(result, line)
			continue
		}

		// Parse timestamp [mm:ss.xx]
		if len(line) >= 10 && line[0] == '[' {
			// Extract timestamp
			endBracket := strings.Index(line, "]")
			if endBracket == -1 {
				result = append(result, line)
				continue
			}

			timestamp := line[1:endBracket]
			text := line[endBracket+1:]

			newTimestamp := adjustTimestamp(timestamp, offsetMs)
			result = append(result, fmt.Sprintf("[%s]%s", newTimestamp, text))
		} else {
			result = append(result, line)
		}
	}

	return strings.Join(result, "\n")
}

func isTimestampLine(line string) bool {
	// A timestamp line starts with [mm:ss.xx] or [mm:ss:xx]
	if len(line) < 10 || line[0] != '[' {
		return false
	}
	// Check if it looks like a timestamp (has : in the right place)
	return (line[3] == ':' || line[3] == '.') && (line[6] == '.' || line[6] == ':')
}

func adjustTimestamp(timestamp string, offsetMs int) string {
	// Parse mm:ss.xx format
	parts := strings.Split(timestamp, ":")
	if len(parts) != 2 {
		return timestamp
	}

	var minutes, seconds int
	var centiseconds int

	// Parse minutes
	fmt.Sscanf(parts[0], "%d", &minutes)

	// Parse seconds and centiseconds
	secParts := strings.Split(parts[1], ".")
	fmt.Sscanf(secParts[0], "%d", &seconds)
	if len(secParts) > 1 {
		fmt.Sscanf(secParts[1], "%d", &centiseconds)
	}

	// Convert to total milliseconds
	totalMs := (minutes*60+seconds)*1000 + centiseconds*10
	totalMs += offsetMs

	// Clamp to 0
	if totalMs < 0 {
		totalMs = 0
	}

	// Convert back
	minutes = totalMs / 60000
	seconds = (totalMs % 60000) / 1000
	centiseconds = (totalMs % 1000) / 10

	return fmt.Sprintf("%02d:%02d.%02d", minutes, seconds, centiseconds)
}
