// Package parser parses subtitle files into structured blocks.
package parser

import (
	"bufio"
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"
)

// Block represents a single subtitle entry.
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
			if b, ok := parseSRTBlock(raw); ok {
				blocks = append(blocks, b)
			}
			raw = nil
		} else {
			raw = append(raw, line)
		}
	}
	if b, ok := parseSRTBlock(raw); ok {
		blocks = append(blocks, b)
	}
	return blocks, scanner.Err()
}

func parseSRTBlock(lines []string) (Block, bool) {
	if len(lines) < 3 {
		return Block{}, false
	}
	seq, err := strconv.Atoi(strings.TrimSpace(lines[0]))
	if err != nil {
		return Block{}, false
	}
	start, end, err := parseSRTTimestampLine(lines[1])
	if err != nil {
		return Block{}, false
	}
	return Block{Sequence: seq, Start: start, End: end, Lines: lines[2:]}, true
}

func parseSRTTimestampLine(s string) (start, end time.Duration, err error) {
	parts := strings.SplitN(s, " --> ", 2)
	if len(parts) != 2 {
		return 0, 0, fmt.Errorf("invalid timestamp line: %q", s)
	}
	start, err = parseTime(strings.TrimSpace(parts[0]), ',')
	if err != nil {
		return
	}
	end, err = parseTime(strings.TrimSpace(parts[1]), ',')
	return
}
