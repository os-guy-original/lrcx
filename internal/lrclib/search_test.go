package lrclib

import (
	"strings"
	"testing"
)

func TestApplyOffset(t *testing.T) {
	tests := []struct {
		name     string
		lrc      string
		offsetMs int
		want     string
	}{
		{
			name:     "positive offset",
			lrc:      "[00:00.00]Hello\n[00:05.00]World",
			offsetMs: 1000,
			want:     "[00:01.00]Hello\n[00:06.00]World",
		},
		{
			name:     "negative offset",
			lrc:      "[00:10.00]Hello",
			offsetMs: -5000,
			want:     "[00:05.00]Hello",
		},
		{
			name:     "zero offset",
			lrc:      "[00:00.00]Test",
			offsetMs: 0,
			want:     "[00:00.00]Test",
		},
		{
			name:     "negative offset clamped to zero",
			lrc:      "[00:00.50]Hello",
			offsetMs: -1000,
			want:     "[00:00.00]Hello",
		},
		{
			name:     "empty string",
			lrc:      "",
			offsetMs: 1000,
			want:     "",
		},
		{
			name:     "skip metadata lines",
			lrc:      "[ti:Song Title]\n[ar:Artist]\n[00:00.00]Lyrics",
			offsetMs: 1000,
			want:     "[ti:Song Title]\n[ar:Artist]\n[00:01.00]Lyrics",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := applyOffset(tt.lrc, tt.offsetMs)
			if got != tt.want {
				t.Errorf("applyOffset() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestIsTimestampLine(t *testing.T) {
	tests := []struct {
		line string
		want bool
	}{
		{"[00:00.00]Hello", true},
		{"[01:30.50]World", true},
		{"[ti:Title]", false},
		{"[ar:Artist]", false},
		{"plain text", false},
		{"", false},
		{"[00:00]", false}, // too short
	}

	for _, tt := range tests {
		t.Run(tt.line, func(t *testing.T) {
			if got := isTimestampLine(tt.line); got != tt.want {
				t.Errorf("isTimestampLine(%q) = %v, want %v", tt.line, got, tt.want)
			}
		})
	}
}

func TestAdjustTimestamp(t *testing.T) {
	tests := []struct {
		timestamp string
		offsetMs  int
		want      string
	}{
		{"00:00.00", 1000, "00:01.00"},
		{"00:00.00", 500, "00:00.50"},
		{"00:05.00", -1000, "00:04.00"},
		{"01:30.00", 30000, "02:00.00"},
		{"00:00.10", -100, "00:00.00"}, // clamped to 0
		{"05:30.50", -330500, "00:00.00"}, // clamped to 0
		{"invalid", 1000, "invalid"}, // invalid format preserved
	}

	for _, tt := range tests {
		t.Run(tt.timestamp, func(t *testing.T) {
			got := adjustTimestamp(tt.timestamp, tt.offsetMs)
			if got != tt.want {
				t.Errorf("adjustTimestamp(%q, %d) = %q, want %q", tt.timestamp, tt.offsetMs, got, tt.want)
			}
		})
	}
}

func TestRunWithOpts_InvalidQuery(t *testing.T) {
	opts := Options{} // Empty query
	err := RunWithOpts(opts)
	if err == nil {
		t.Error("expected error for empty query")
	}
	if !strings.Contains(err.Error(), "query") {
		t.Errorf("unexpected error message: %v", err)
	}
}
