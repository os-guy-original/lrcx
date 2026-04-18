// Package parser parses SRT subtitle files into structured blocks.
package parser

import (
	"bufio"
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"
)

// Block represents a single SRT subtitle entry.
type Block struct {
	Sequence int
	Start    time.Duration
	End      time.Duration
	Lines    []string
}

// ParseSRT reads an SRT stream and returns all valid blocks.
func ParseSRT(r io.Reader) ([]Block, error) {
	var blocks []Block
	var raw []string

	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		line := strings.TrimRight(scanner.Text(), "\r")
		if line == "" {
			if b, ok := parseBlock(raw); ok {
				blocks = append(blocks, b)
			}
			raw = nil
		} else {
			raw = append(raw, line)
		}
	}
	if len(raw) > 0 {
		if b, ok := parseBlock(raw); ok {
			blocks = append(blocks, b)
		}
	}
	return blocks, scanner.Err()
}

func parseBlock(lines []string) (Block, bool) {
	if len(lines) < 3 {
		return Block{}, false
	}
	seq, err := strconv.Atoi(strings.TrimSpace(lines[0]))
	if err != nil {
		return Block{}, false
	}
	start, end, err := parseTimestampLine(lines[1])
	if err != nil {
		return Block{}, false
	}
	return Block{
		Sequence: seq,
		Start:    start,
		End:      end,
		Lines:    lines[2:],
	}, true
}

// parseTimestampLine parses "HH:MM:SS,mmm --> HH:MM:SS,mmm".
func parseTimestampLine(s string) (start, end time.Duration, err error) {
	parts := strings.SplitN(s, " --> ", 2)
	if len(parts) != 2 {
		return 0, 0, fmt.Errorf("invalid timestamp line: %q", s)
	}
	start, err = parseSRTTime(strings.TrimSpace(parts[0]))
	if err != nil {
		return
	}
	end, err = parseSRTTime(strings.TrimSpace(parts[1]))
	return
}

// parseSRTTime parses "HH:MM:SS,mmm" into a duration.
func parseSRTTime(s string) (time.Duration, error) {
	// Normalize comma to dot for milliseconds.
	s = strings.ReplaceAll(s, ",", ".")
	var h, m, sec, ms int
	_, err := fmt.Sscanf(s, "%d:%d:%d.%d", &h, &m, &sec, &ms)
	if err != nil {
		return 0, fmt.Errorf("invalid SRT time %q: %w", s, err)
	}
	d := time.Duration(h)*time.Hour +
		time.Duration(m)*time.Minute +
		time.Duration(sec)*time.Second +
		time.Duration(ms)*time.Millisecond
	return d, nil
}
