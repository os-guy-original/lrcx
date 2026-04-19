package parser

import (
	"fmt"
	"strings"
	"time"
)

// parseTime parses a subtitle timestamp with the given millisecond separator (',' for SRT, '.' for VTT).
// Accepts "HH:MM:SS<sep>mmm" or "MM:SS<sep>mmm".
func parseTime(s string, sep rune) (time.Duration, error) {
	s = strings.ReplaceAll(s, string(sep), ".")
	var h, m, sec, ms int
	var err error
	if strings.Count(s, ":") == 2 {
		_, err = fmt.Sscanf(s, "%d:%d:%d.%d", &h, &m, &sec, &ms)
	} else {
		_, err = fmt.Sscanf(s, "%d:%d.%d", &m, &sec, &ms)
	}
	if err != nil {
		return 0, fmt.Errorf("invalid time %q: %w", s, err)
	}
	return time.Duration(h)*time.Hour +
		time.Duration(m)*time.Minute +
		time.Duration(sec)*time.Second +
		time.Duration(ms)*time.Millisecond, nil
}
