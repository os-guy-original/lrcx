// Package converter converts parsed SRT blocks to LRC lines.
package converter

import (
	"fmt"
	"strings"
	"time"

	"github.com/srt2lrc/srt2lrc/internal/parser"
	"github.com/srt2lrc/srt2lrc/internal/tags"
)

// ToLRC converts SRT blocks to LRC-formatted lines.
// offset shifts all start times (can be negative).
func ToLRC(blocks []parser.Block, offset time.Duration) []string {
	out := make([]string, 0, len(blocks))
	for _, b := range blocks {
		text := tags.Process(strings.Join(b.Lines, " "))
		text = strings.TrimSpace(text)
		if text == "" {
			continue
		}
		start := b.Start + offset
		if start < 0 {
			start = 0
		}
		out = append(out, fmt.Sprintf("%s%s", formatLRCTime(start), text))
	}
	return out
}

// formatLRCTime formats a duration as [mm:ss.cc].
func formatLRCTime(d time.Duration) string {
	total := d.Milliseconds()
	cs := (total % 1000) / 10
	total /= 1000
	sec := total % 60
	min := total / 60
	return fmt.Sprintf("[%02d:%02d.%02d]", min, sec, cs)
}
