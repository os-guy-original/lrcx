// Package parser parses VTT subtitle files into structured blocks.
package parser

import (
	"bufio"
	"fmt"
	"io"
	"strings"
	"time"
)

// ParseVTT reads a VTT stream and returns all valid blocks.
func ParseVTT(r io.Reader) ([]Block, error) {
	var blocks []Block
	var raw []string

	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		line := strings.TrimRight(scanner.Text(), "\r")

		// Skip WEBVTT header and empty lines at start
		if strings.HasPrefix(line, "WEBVTT") {
			continue
		}

		// Empty line ends a block
		if line == "" {
			if len(raw) >= 2 {
				if b, ok := parseVTTBlock(raw); ok {
					blocks = append(blocks, b)
				}
			}
			raw = nil
			continue
		}

		raw = append(raw, line)
	}

	// Handle last block
	if len(raw) >= 2 {
		if b, ok := parseVTTBlock(raw); ok {
			blocks = append(blocks, b)
		}
	}

	return blocks, scanner.Err()
}

func parseVTTBlock(lines []string) (Block, bool) {
	if len(lines) < 2 {
		return Block{}, false
	}

	// First line is timestamp
	start, end, err := parseVTTTimestampLine(lines[0])
	if err != nil {
		return Block{}, false
	}

	// Rest are text lines
	textLines := lines[1:]
	// Filter out cue settings (lines starting with position:)
	var filtered []string
	for _, l := range textLines {
		if !strings.HasPrefix(l, "position:") && !strings.HasPrefix(l, "align:") {
			filtered = append(filtered, l)
		}
	}

	return Block{
		Sequence: 0,
		Start:    start,
		End:      end,
		Lines:    filtered,
	}, true
}

// parseVTTTimestampLine parses "HH:MM:SS.mmm --> HH:MM:SS.mmm" or "MM:SS.mmm --> MM:SS.mmm".
func parseVTTTimestampLine(s string) (start, end time.Duration, err error) {
	parts := strings.SplitN(s, " --> ", 2)
	if len(parts) != 2 {
		return 0, 0, fmt.Errorf("invalid timestamp line: %q", s)
	}

	// Remove cue settings from end timestamp (e.g., "00:00:04.000 position:50%")
	endPart := parts[1]
	if idx := strings.Index(endPart, " "); idx > 0 {
		endPart = endPart[:idx]
	}

	start, err = parseVTTTime(strings.TrimSpace(parts[0]))
	if err != nil {
		return
	}
	end, err = parseVTTTime(strings.TrimSpace(endPart))
	return
}

// parseVTTTime parses "HH:MM:SS.mmm" or "MM:SS.mmm" into a duration.
func parseVTTTime(s string) (time.Duration, error) {
	var h, m, sec, ms int

	// Split at dot for milliseconds
	timePart := s
	msPart := "0"
	if idx := strings.Index(s, "."); idx >= 0 {
		timePart = s[:idx]
		msPart = s[idx+1:]
	}

	// Parse milliseconds (pad to 3 digits)
	if len(msPart) > 3 {
		msPart = msPart[:3]
	}
	for len(msPart) < 3 {
		msPart += "0"
	}
	_, err := fmt.Sscanf(msPart, "%d", &ms)
	if err != nil {
		return 0, fmt.Errorf("invalid VTT time %q: %w", s, err)
	}

	// Parse time part (HH:MM:SS or MM:SS)
	parts := strings.Split(timePart, ":")
	if len(parts) == 3 {
		_, err = fmt.Sscanf(timePart, "%d:%d:%d", &h, &m, &sec)
	} else if len(parts) == 2 {
		_, err = fmt.Sscanf(timePart, "%d:%d", &m, &sec)
	} else {
		return 0, fmt.Errorf("invalid VTT time format: %q", s)
	}
	if err != nil {
		return 0, fmt.Errorf("invalid VTT time %q: %w", s, err)
	}

	d := time.Duration(h)*time.Hour +
		time.Duration(m)*time.Minute +
		time.Duration(sec)*time.Second +
		time.Duration(ms)*time.Millisecond
	return d, nil
}
