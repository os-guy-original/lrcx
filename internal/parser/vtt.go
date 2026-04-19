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

	flush := func() {
		if b, ok := parseVTTBlock(raw); ok {
			blocks = append(blocks, b)
		}
		raw = nil
	}

	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		line := strings.TrimRight(scanner.Text(), "\r")
		if strings.HasPrefix(line, "WEBVTT") || strings.HasPrefix(line, "Kind:") || strings.HasPrefix(line, "Language:") {
			continue
		}
		if line == "" {
			flush()
		} else {
			raw = append(raw, line)
		}
	}
	flush()

	return blocks, scanner.Err()
}

func parseVTTBlock(lines []string) (Block, bool) {
	if len(lines) < 2 {
		return Block{}, false
	}
	// Skip optional cue identifier (non-timestamp first line)
	start := 0
	if !strings.Contains(lines[0], "-->") {
		start = 1
	}
	if start >= len(lines)-1 {
		return Block{}, false
	}
	s, e, err := parseVTTTimestampLine(lines[start])
	if err != nil {
		return Block{}, false
	}
	return Block{Start: s, End: e, Lines: lines[start+1:]}, true
}

func parseVTTTimestampLine(s string) (start, end time.Duration, err error) {
	parts := strings.SplitN(s, " --> ", 2)
	if len(parts) != 2 {
		return 0, 0, fmt.Errorf("invalid timestamp line: %q", s)
	}
	// Strip cue settings from end (e.g. "00:04.000 position:50%")
	endPart := parts[1]
	if i := strings.IndexByte(endPart, ' '); i > 0 {
		endPart = endPart[:i]
	}
	start, err = parseTime(strings.TrimSpace(parts[0]), '.')
	if err != nil {
		return
	}
	end, err = parseTime(endPart, '.')
	return
}
